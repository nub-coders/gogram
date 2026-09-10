// Copyright (c) 2025 @AmarnathCJD

package telegram

import (
	"fmt"
)

// EphemeralOptions configures an ephemeral message: one that is delivered into
// a group but rendered only for a single receiving user and the bot.
//
// It mirrors the Bot API 10.3 EphemeralMessageParameters object, which replaced
// the standalone receiver_user_id and callback_query_id parameters.
type EphemeralOptions struct {
	// Peer is the chat the ephemeral message appears in. Leave nil to send
	// into the private chat with the receiver.
	Peer any

	// QueryID ties the message to a callback query. When set together with
	// ReplaceCallbackQueryMessage, the ephemeral message is shown in place of
	// the message that carried the button.
	QueryID int64

	// ReplaceCallbackQueryMessage renders this message in place of the
	// original message instead of appending a new one. Requires QueryID.
	ReplaceCallbackQueryMessage bool

	// Welcome marks the message as a welcome message, shown to users when
	// they first open the chat. Requires the manage_welcome_messages right.
	Welcome bool

	// Anchor pins the message to the bottom of the chat for the receiver.
	Anchor bool

	InvertMedia            bool // Display media below the caption instead of above
	NoForwards             bool // Restrict forwarding and saving
	// ShowCaptionAboveMedia renders the caption above the attached media.
	// Applies only to editEphemeralMessageCaption (Bot API 10.3).
	ShowCaptionAboveMedia bool

	Entities    []MessageEntity
	Media       InputMedia
	ReplyMarkup ReplyMarkup
	ReplyTo     InputReplyTo

	// Rich renders the message as a rich message. When set, Text is ignored.
	Rich *RichBuilder
}

func (c *Client) ephemeralPeer(o *EphemeralOptions) (InputPeer, error) {
	if o == nil || o.Peer == nil {
		return nil, nil
	}

	peer, err := c.GetSendablePeer(o.Peer)
	if err != nil {
		return nil, fmt.Errorf("resolve ephemeral peer: %w", err)
	}

	return peer, nil
}

// SendEphemeral sends a message visible only to receiver inside the given chat.
//
// The receiver may be a user ID, username, or InputUser. Pass options to target
// a specific group, attach media, or answer a callback query in place:
//
//	client.SendEphemeral(userID, "Only you can see this", &telegram.EphemeralOptions{
//		Peer: groupID,
//	})
//
// The returned EphemeralMessageEvent wraps the server-sent ephemeral message and
// exposes Edit/Delete helpers. It is nil only when err is non-nil.
func (c *Client) SendEphemeral(receiver any, text string, opts ...*EphemeralOptions) (*EphemeralMessageEvent, error) {
	o := getVariadic(opts, &EphemeralOptions{})

	user, err := c.GetSendableUser(receiver)
	if err != nil {
		return nil, fmt.Errorf("resolve ephemeral receiver: %w", err)
	}

	peer, err := c.ephemeralPeer(o)
	if err != nil {
		return nil, err
	}

	if o.ReplaceCallbackQueryMessage && o.QueryID == 0 {
		return nil, fmt.Errorf("ephemeral: ReplaceCallbackQueryMessage requires QueryID")
	}

	params := &EphemeralSendMessageParams{
		InvertMedia: o.InvertMedia,
		Welcome:     o.Welcome,
		Anchor:      o.Anchor || o.ReplaceCallbackQueryMessage,
		Noforwards:  o.NoForwards,
		Peer:        peer,
		ReceiverID:  user,
		QueryID:     o.QueryID,
		Message:     text,
		Entities:    o.Entities,
		Media:       o.Media,
		ReplyMarkup: o.ReplyMarkup,
		RandomID:    GenRandInt(),
		ReplyTo:     o.ReplyTo,
	}

	if o.Rich != nil {
		if err := o.Rich.resolve(c); err != nil {
			return nil, err
		}

		params.RichMessage = o.Rich.build()
		params.Message = ""
	}

	resp, err := c.EphemeralSendMessage(params)
	if err != nil {
		return nil, err
	}

	message, raw, err := extractEphemeralMessage(resp)
	if err != nil {
		return nil, err
	}
	if peer != nil && message.PeerID == nil {
		message.PeerID = c.getPeer(peer)
	}

	return packEphemeralMessage(c, message, raw), nil
}

// EditEphemeral edits a previously sent ephemeral message.
func (c *Client) EditEphemeral(receiver any, messageID int32, text string, opts ...*EphemeralOptions) (*EphemeralMessageEvent, error) {
	o := getVariadic(opts, &EphemeralOptions{})

	user, err := c.GetSendableUser(receiver)
	if err != nil {
		return nil, fmt.Errorf("resolve ephemeral receiver: %w", err)
	}

	peer, err := c.ephemeralPeer(o)
	if err != nil {
		return nil, err
	}

	params := &EphemeralEditMessageParams{
		InvertMedia:           o.InvertMedia,
		Welcome:               o.Welcome,
		ShowCaptionAboveMedia: o.ShowCaptionAboveMedia,
		Peer:                  peer,
		ReceiverID:            user,
		ID:                    messageID,
		Message:               text,
		Media:                 o.Media,
		Entities:              o.Entities,
		ReplyMarkup:           o.ReplyMarkup,
	}

	if o.Rich != nil {
		if err := o.Rich.resolve(c); err != nil {
			return nil, err
		}

		params.RichMessage = o.Rich.build()
		params.Message = ""
	}

	resp, err := c.EphemeralEditMessage(params)
	if err != nil {
		return nil, err
	}

	message, raw, err := extractEphemeralMessage(resp)
	if err != nil {
		return nil, err
	}
	if peer != nil && message.PeerID == nil {
		message.PeerID = c.getPeer(peer)
	}

	return packEphemeralMessage(c, message, raw), nil
}

// extractEphemeralMessage pulls the ephemeral message out of an ephemeral
// send/edit RPC response. Unlike ordinary messages, ephemeral.sendMessage and
// ephemeral.editMessage return updateNewEphemeralMessage / updateEditEphemeralMessage
// (never updateNewMessage), so they must be unpacked here rather than via
// processUpdate, which only understands ordinary message updates.
func extractEphemeralMessage(resp Updates) (*EphemeralMessage, Update, error) {
	if resp == nil {
		return nil, nil, fmt.Errorf("no response from server")
	}

	var updates []Update
	switch u := resp.(type) {
	case *UpdatesObj:
		updates = u.Updates
	case *UpdatesCombined:
		updates = u.Updates
	case *UpdateShort:
		updates = []Update{u.Update}
	default:
		return nil, nil, fmt.Errorf("unexpected ephemeral response type %T", resp)
	}

	for _, upd := range updates {
		switch u := upd.(type) {
		case *UpdateNewEphemeralMessage:
			if u.Message != nil {
				return u.Message, u, nil
			}
		case *UpdateEditEphemeralMessage:
			if u.Message != nil {
				return u.Message, u, nil
			}
		}
	}

	return nil, nil, fmt.Errorf("no ephemeral message update in server response")
}

// DeleteEphemeral deletes an ephemeral message previously sent to receiver.
func (c *Client) DeleteEphemeral(receiver any, messageID int32, peerID ...any) (bool, error) {
	user, err := c.GetSendableUser(receiver)
	if err != nil {
		return false, fmt.Errorf("resolve ephemeral receiver: %w", err)
	}

	var peer InputPeer
	if len(peerID) > 0 && peerID[0] != nil {
		if peer, err = c.GetSendablePeer(peerID[0]); err != nil {
			return false, fmt.Errorf("resolve ephemeral peer: %w", err)
		}
	}

	return c.EphemeralDeleteMessage(peer, user, messageID)
}

// SendWelcomeMessage sends an ephemeral welcome message shown to users when
// they open the chat. Requires the manage_welcome_messages admin right.
func (c *Client) SendWelcomeMessage(peerID any, receiver any, text string, opts ...*EphemeralOptions) (*EphemeralMessageEvent, error) {
	o := getVariadic(opts, &EphemeralOptions{})
	o.Welcome = true
	o.Peer = peerID

	return c.SendEphemeral(receiver, text, o)
}

// GetWelcomeMessages returns the welcome messages configured for a chat.
func (c *Client) GetWelcomeMessages(peerID any, hash ...int64) (EphemeralWelcomeMessages, error) {
	peer, err := c.GetSendablePeer(peerID)
	if err != nil {
		return nil, fmt.Errorf("resolve peer: %w", err)
	}

	return c.EphemeralGetWelcomeMessages(peer, getVariadic(hash, int64(0)))
}

// DeleteWelcomeMessage removes a single welcome message from a chat.
func (c *Client) DeleteWelcomeMessage(peerID any, messageID int32) (bool, error) {
	peer, err := c.GetSendablePeer(peerID)
	if err != nil {
		return false, fmt.Errorf("resolve peer: %w", err)
	}

	return c.EphemeralDeleteWelcomeMessage(peer, messageID)
}

// DeleteAllWelcomeMessages removes every welcome message from a chat.
func (c *Client) DeleteAllWelcomeMessages(peerID any) (bool, error) {
	peer, err := c.GetSendablePeer(peerID)
	if err != nil {
		return false, fmt.Errorf("resolve peer: %w", err)
	}

	return c.EphemeralDeleteAllWelcomeMessages(peer)
}

// GetEphemeralCallbackAnswer presses an inline button on an ephemeral message
// and returns the bot's answer. data is the button payload; pass nil for
// buttons that carry none.
func (c *Client) GetEphemeralCallbackAnswer(peerID any, messageID int32, data []byte) (*MessagesBotCallbackAnswer, error) {
	peer, err := c.GetSendablePeer(peerID)
	if err != nil {
		return nil, fmt.Errorf("resolve peer: %w", err)
	}

	return c.EphemeralGetCallbackAnswer(peer, messageID, data)
}

// ReportEphemeral reports an ephemeral message. option selects a node in the
// server-driven report menu; pass nil to request the top-level options, then
// re-call with the chosen option until a ReportResultReported is returned.
func (c *Client) ReportEphemeral(peerID any, messageID int32, option []byte, message string) (ReportResult, error) {
	peer, err := c.GetSendablePeer(peerID)
	if err != nil {
		return nil, fmt.Errorf("resolve peer: %w", err)
	}

	return c.EphemeralReportMessage(peer, messageID, option, message)
}

// EphemeralMessageEvent is the high-level wrapper around a layer-229
// EphemeralMessage. It carries the client and resolved sender/chat so handlers
// and RPC results can act on the message without re-resolving peers.
//
// The embedded *EphemeralMessage exposes every raw field (ID, Message,
// Entities, ReceiverID, ChatInstance, ...).
type EphemeralMessageEvent struct {
	*EphemeralMessage
	Client    *Client
	Sender    *UserObj
	Chat      *ChatObj
	Channel   *Channel
	RawUpdate Update
}

// Text returns the plain message text of the ephemeral message.
func (m *EphemeralMessageEvent) Text() string {
	if m == nil || m.EphemeralMessage == nil {
		return ""
	}
	return m.EphemeralMessage.Message
}

// ChatID returns the raw peer ID the ephemeral message is shown in.
func (m *EphemeralMessageEvent) ChatID() int64 {
	if m == nil || m.EphemeralMessage == nil {
		return 0
	}
	switch peer := m.PeerID.(type) {
	case *PeerUser:
		return peer.UserID
	case *PeerChat:
		return peer.ChatID
	case *PeerChannel:
		return peer.ChannelID
	}
	return m.ReceiverID
}

// SenderID returns the ID of the user or peer that sent the message.
func (m *EphemeralMessageEvent) SenderID() int64 {
	if m == nil {
		return 0
	}
	if m.Sender != nil && m.Sender.ID != 0 {
		return m.Sender.ID
	}
	switch from := m.FromID.(type) {
	case *PeerUser:
		return from.UserID
	case *PeerChat:
		return from.ChatID
	case *PeerChannel:
		return from.ChannelID
	}
	return 0
}

func (m *EphemeralMessageEvent) IsPrivate() bool {
	if m == nil || m.EphemeralMessage == nil {
		return false
	}
	if m.PeerID == nil {
		return true // targeted directly at the receiver's private chat
	}
	_, ok := m.PeerID.(*PeerUser)
	return ok
}

func (m *EphemeralMessageEvent) IsGroup() bool {
	if m == nil || m.EphemeralMessage == nil {
		return false
	}
	_, ok := m.PeerID.(*PeerChat)
	return ok
}

func (m *EphemeralMessageEvent) IsChannel() bool {
	if m == nil || m.EphemeralMessage == nil {
		return false
	}
	_, ok := m.PeerID.(*PeerChannel)
	return ok
}

func (m *EphemeralMessageEvent) IsOutgoing() bool {
	return m != nil && m.EphemeralMessage != nil && m.Out
}

func (m *EphemeralMessageEvent) Marshal(noindent ...bool) string {
	if m == nil {
		return "null"
	}
	if m.RawUpdate != nil {
		return MarshalWithTypeName(m.RawUpdate, noindent...)
	}
	return MarshalWithTypeName(m.EphemeralMessage, noindent...)
}

// Edit edits this ephemeral message with new text.
func (m *EphemeralMessageEvent) Edit(text string, opts ...*EphemeralOptions) (*EphemeralMessageEvent, error) {
	if m == nil || m.Client == nil {
		return nil, fmt.Errorf("ephemeral message client unavailable")
	}
	o := getVariadic(opts, &EphemeralOptions{})
	if o.Peer == nil && m.PeerID != nil {
		o.Peer = m.PeerID
	}
	return m.Client.EditEphemeral(m.ReceiverID, m.ID, text, o)
}

// Delete deletes this ephemeral message.
func (m *EphemeralMessageEvent) Delete() (bool, error) {
	if m == nil || m.Client == nil {
		return false, fmt.Errorf("ephemeral message client unavailable")
	}
	if m.PeerID != nil {
		return m.Client.DeleteEphemeral(m.ReceiverID, m.ID, m.PeerID)
	}
	return m.Client.DeleteEphemeral(m.ReceiverID, m.ID)
}

// packEphemeralMessage wraps a raw EphemeralMessage with the client and
// resolved sender/chat metadata. rawUpdate, when non-nil, is the update the
// message arrived in (used for faithful Marshal output).
func packEphemeralMessage(c *Client, message *EphemeralMessage, rawUpdate Update) *EphemeralMessageEvent {
	if message == nil {
		return nil
	}
	packed := &EphemeralMessageEvent{
		EphemeralMessage: message,
		Client:           c,
		RawUpdate:        rawUpdate,
	}
	if message.FromID != nil {
		packed.Sender = c.getSender(message.FromID)
	}
	packed.Chat = c.getChat(message.PeerID)
	packed.Channel = c.getChannel(message.PeerID)
	if packed.Chat == nil && packed.Sender != nil && packed.Sender.ID != 0 {
		packed.Chat = &ChatObj{ID: packed.Sender.ID, Title: packed.Sender.FirstName}
	}
	return packed
}

// EphemeralDeleteMessage represents a batch deletion of ephemeral messages
// (updateDeleteEphemeralMessages).
type EphemeralDeleteMessage struct {
	Client    *Client
	ChannelID int64
	Messages  []int32
	Peer      Peer
	RawUpdate Update
}

func (m *EphemeralDeleteMessage) Marshal(noindent ...bool) string {
	if m == nil || m.RawUpdate == nil {
		return "null"
	}
	return MarshalWithTypeName(m.RawUpdate, noindent...)
}

func packEphemeralDelete(c *Client, update *UpdateDeleteEphemeralMessages) *EphemeralDeleteMessage {
	if update == nil {
		return nil
	}
	return &EphemeralDeleteMessage{
		Client:    c,
		ChannelID: c.GetPeerID(update.Peer),
		Messages:  update.Ids,
		Peer:      update.Peer,
		RawUpdate: update,
	}
}

// EphemeralCallbackQuery represents an ephemeral inline button press
// (updateEphemeralBotCallbackQuery).
type EphemeralCallbackQuery struct {
	QueryID        int64
	Data           []byte
	OriginalUpdate *UpdateEphemeralBotCallbackQuery
	Sender         *UserObj
	MessageID      int32
	SenderID       int64
	ChatID         int64
	Message        *EphemeralMessageEvent
	ChatInstance   int64
	Client         *Client
	Peer           Peer
	Chat           *ChatObj
	Channel        *Channel
}

func (b *EphemeralCallbackQuery) DataString() string {
	return string(b.Data)
}

// Answer answers the callback query by pressing the button server-side and
// returning the bot's answer. Ephemeral callbacks are answered via
// ephemeral.getCallbackAnswer, not messages.setBotCallbackAnswer.
func (b *EphemeralCallbackQuery) Answer() (*MessagesBotCallbackAnswer, error) {
	if b == nil || b.Client == nil {
		return nil, fmt.Errorf("ephemeral callback client unavailable")
	}
	if b.Peer == nil {
		return nil, fmt.Errorf("ephemeral callback has no peer to answer")
	}
	return b.Client.GetEphemeralCallbackAnswer(b.Peer, b.MessageID, b.Data)
}

// Edit edits the message that carried the pressed button.
func (b *EphemeralCallbackQuery) Edit(text string, opts ...*EphemeralOptions) (*EphemeralMessageEvent, error) {
	if b == nil || b.Client == nil || b.Message == nil {
		return nil, fmt.Errorf("ephemeral callback message unavailable")
	}
	return b.Message.Edit(text, opts...)
}

// Delete deletes the message that carried the pressed button.
func (b *EphemeralCallbackQuery) Delete() (bool, error) {
	if b == nil || b.Client == nil || b.Message == nil {
		return false, fmt.Errorf("ephemeral callback message unavailable")
	}
	return b.Message.Delete()
}

func (b *EphemeralCallbackQuery) ChatType() string {
	if b == nil || b.Peer == nil {
		return EntityUnknown
	}
	switch b.Peer.(type) {
	case *PeerUser:
		return EntityUser
	case *PeerChat:
		return EntityChat
	case *PeerChannel:
		return EntityChannel
	}
	return EntityUnknown
}

func (b *EphemeralCallbackQuery) IsPrivate() bool {
	return b.ChatType() == EntityUser
}

func (b *EphemeralCallbackQuery) IsGroup() bool {
	if b.Channel != nil {
		return b.ChatType() == EntityChat || (b.ChatType() == EntityChannel && !b.Channel.Broadcast)
	}
	return b.ChatType() == EntityChat
}

func (b *EphemeralCallbackQuery) IsChannel() bool {
	return b.ChatType() == EntityChannel
}

func (b *EphemeralCallbackQuery) Marshal(noindent ...bool) string {
	if b == nil || b.OriginalUpdate == nil {
		return "null"
	}
	return MarshalWithTypeName(b.OriginalUpdate, noindent...)
}

func packEphemeralCallback(c *Client, update *UpdateEphemeralBotCallbackQuery) *EphemeralCallbackQuery {
	if update == nil {
		return nil
	}
	cq := &EphemeralCallbackQuery{
		QueryID:        update.QueryID,
		Data:           update.Data,
		OriginalUpdate: update,
		Client:         c,
		Sender:         c.getSender(&PeerUser{UserID: update.UserID}),
		MessageID:      update.MsgID,
		SenderID:       update.UserID,
		ChatInstance:   update.ChatInstance,
		Message:        packEphemeralMessage(c, update.Message, update),
	}
	if update.Peer != nil {
		cq.Peer = update.Peer
	} else if update.Message != nil {
		cq.Peer = update.Message.PeerID
	}
	cq.ChatID = c.GetPeerID(cq.Peer)
	cq.Chat = c.getChat(cq.Peer)
	cq.Channel = c.getChannel(cq.Peer)
	return cq
}
