// Copyright (c) 2025 @AmarnathCJD

package telegram

import (
	"errors"
	"fmt"
	"maps"
)

// This file completes dispatch coverage for the bot-facing updates introduced
// up to Bot API 10.2/10.3: business connections and their messages, payment
// queries, reaction updates, poll votes, boosts, paid media purchases, star
// subscriptions and basic-group participant changes.
//
// Every update already reached raw handlers via handleRawUpdate; these types
// add the typed, peer-resolved access that the older events (messages,
// callbacks, inline queries) already had.

// ---------------------------- Event Types ----------------------------

// BusinessMessage is a message received or edited in a chat the bot is
// connected to through a Telegram Business connection.
type BusinessMessage struct {
	Client         *Client
	OriginalUpdate Update
	// ConnectionID identifies the business connection that delivered this
	// message; pass it back when replying so the message is sent on behalf
	// of the business account.
	ConnectionID string
	Message      *NewMessage
	ReplyTo      *NewMessage
	Qts          int32
}

// Text returns the message text, if any.
func (b *BusinessMessage) Text() string {
	if b == nil || b.Message == nil {
		return ""
	}
	return b.Message.Text()
}

// ChatID returns the chat the business message belongs to.
func (b *BusinessMessage) ChatID() int64 {
	if b == nil || b.Message == nil {
		return 0
	}
	return b.Message.ChatID()
}

// SenderID returns the user that sent the business message.
func (b *BusinessMessage) SenderID() int64 {
	if b == nil || b.Message == nil {
		return 0
	}
	return b.Message.SenderID()
}

func (b *BusinessMessage) Marshal(noindent ...bool) string {
	if b == nil || b.OriginalUpdate == nil {
		return "null"
	}
	return MarshalWithTypeName(b.OriginalUpdate, noindent...)
}

// BusinessMessageDeleted reports messages deleted in a connected business chat.
type BusinessMessageDeleted struct {
	Client         *Client
	OriginalUpdate *UpdateBotDeleteBusinessMessage
	ConnectionID   string
	Peer           Peer
	ChatID         int64
	Messages       []int32
	Qts            int32
}

func (b *BusinessMessageDeleted) Marshal(noindent ...bool) string {
	if b == nil || b.OriginalUpdate == nil {
		return "null"
	}
	return MarshalWithTypeName(b.OriginalUpdate, noindent...)
}

// BusinessConnection reports that a business account connected the bot, or
// changed the rights granted to it.
type BusinessConnection struct {
	Client         *Client
	OriginalUpdate *UpdateBotBusinessConnect
	Connection     *BotBusinessConnection
	ConnectionID   string
	UserID         int64
	Qts            int32
}

// Disabled reports whether the connection is currently disabled.
func (b *BusinessConnection) Disabled() bool {
	if b == nil || b.Connection == nil {
		return false
	}
	return b.Connection.Disabled
}

// Rights returns the rights granted to the bot over the business account.
func (b *BusinessConnection) Rights() *BusinessBotRights {
	if b == nil || b.Connection == nil {
		return nil
	}
	return b.Connection.Rights
}

func (b *BusinessConnection) Marshal(noindent ...bool) string {
	if b == nil || b.OriginalUpdate == nil {
		return "null"
	}
	return MarshalWithTypeName(b.OriginalUpdate, noindent...)
}

// PreCheckoutQuery is a final payment confirmation request. The bot must
// answer within 10 seconds using Ok or Fail, otherwise the payment fails.
type PreCheckoutQuery struct {
	Client           *Client
	OriginalUpdate   *UpdateBotPrecheckoutQuery
	QueryID          int64
	UserID           int64
	Sender           *UserObj
	Payload          []byte
	Info             *PaymentRequestedInfo
	ShippingOptionID string
	Currency         string
	TotalAmount      int64
}

// PayloadString returns the invoice payload as a string.
func (p *PreCheckoutQuery) PayloadString() string {
	if p == nil {
		return ""
	}
	return string(p.Payload)
}

// Ok approves the pending payment.
func (p *PreCheckoutQuery) Ok() (bool, error) {
	if p == nil || p.Client == nil {
		return false, fmt.Errorf("precheckout client unavailable")
	}
	return p.Client.MessagesSetBotPrecheckoutResults(true, p.QueryID, "")
}

// Fail rejects the pending payment with a human readable reason.
func (p *PreCheckoutQuery) Fail(reason string) (bool, error) {
	if p == nil || p.Client == nil {
		return false, fmt.Errorf("precheckout client unavailable")
	}
	return p.Client.MessagesSetBotPrecheckoutResults(false, p.QueryID, reason)
}

func (p *PreCheckoutQuery) Marshal(noindent ...bool) string {
	if p == nil || p.OriginalUpdate == nil {
		return "null"
	}
	return MarshalWithTypeName(p.OriginalUpdate, noindent...)
}

// ShippingQuery is a request for the shipping options available for an order.
type ShippingQuery struct {
	Client          *Client
	OriginalUpdate  *UpdateBotShippingQuery
	QueryID         int64
	UserID          int64
	Sender          *UserObj
	Payload         []byte
	ShippingAddress *PostAddress
}

// PayloadString returns the invoice payload as a string.
func (s *ShippingQuery) PayloadString() string {
	if s == nil {
		return ""
	}
	return string(s.Payload)
}

// Answer returns the available shipping options to the user.
func (s *ShippingQuery) Answer(options []*ShippingOption) (bool, error) {
	if s == nil || s.Client == nil {
		return false, fmt.Errorf("shipping query client unavailable")
	}
	return s.Client.MessagesSetBotShippingResults(s.QueryID, "", options)
}

// Fail rejects the order with a human readable reason.
func (s *ShippingQuery) Fail(reason string) (bool, error) {
	if s == nil || s.Client == nil {
		return false, fmt.Errorf("shipping query client unavailable")
	}
	return s.Client.MessagesSetBotShippingResults(s.QueryID, reason, nil)
}

func (s *ShippingQuery) Marshal(noindent ...bool) string {
	if s == nil || s.OriginalUpdate == nil {
		return "null"
	}
	return MarshalWithTypeName(s.OriginalUpdate, noindent...)
}

// MessageReactionUpdate reports that a specific user changed their reactions
// to a message. Delivered for private chats and small groups.
type MessageReactionUpdate struct {
	Client         *Client
	OriginalUpdate *UpdateBotMessageReaction
	Peer           Peer
	ChatID         int64
	MessageID      int32
	Date           int32
	Actor          Peer
	ActorID        int64
	Sender         *UserObj
	Chat           *ChatObj
	Channel        *Channel
	OldReactions   []Reaction
	NewReactions   []Reaction
}

// Added returns the reactions present after the change but not before.
func (m *MessageReactionUpdate) Added() []Reaction {
	if m == nil {
		return nil
	}
	return reactionsDifference(m.NewReactions, m.OldReactions)
}

// Removed returns the reactions present before the change but not after.
func (m *MessageReactionUpdate) Removed() []Reaction {
	if m == nil {
		return nil
	}
	return reactionsDifference(m.OldReactions, m.NewReactions)
}

func (m *MessageReactionUpdate) Marshal(noindent ...bool) string {
	if m == nil || m.OriginalUpdate == nil {
		return "null"
	}
	return MarshalWithTypeName(m.OriginalUpdate, noindent...)
}

// reactionKey renders a reaction as a comparable key. Anonymous reaction
// updates in large groups only carry counters, so this is used for the
// per-user diff in MessageReactionUpdate.
func reactionKey(r Reaction) string {
	switch v := r.(type) {
	case *ReactionEmoji:
		return "e:" + v.Emoticon
	case *ReactionCustomEmoji:
		return fmt.Sprintf("c:%d", v.DocumentID)
	case *ReactionPaid:
		return "paid"
	case *ReactionEmpty:
		return "empty"
	default:
		return fmt.Sprintf("%T", r)
	}
}

func reactionsDifference(from, minus []Reaction) []Reaction {
	if len(from) == 0 {
		return nil
	}
	seen := make(map[string]int, len(minus))
	for _, r := range minus {
		seen[reactionKey(r)]++
	}
	var out []Reaction
	for _, r := range from {
		key := reactionKey(r)
		if seen[key] > 0 {
			seen[key]--
			continue
		}
		out = append(out, r)
	}
	return out
}

// MessageReactionCountUpdate reports anonymous reaction counters for a message
// in a large group or channel, where individual reactors are not exposed.
type MessageReactionCountUpdate struct {
	Client         *Client
	OriginalUpdate *UpdateBotMessageReactions
	Peer           Peer
	ChatID         int64
	MessageID      int32
	Date           int32
	Chat           *ChatObj
	Channel        *Channel
	Reactions      []*ReactionCount
}

func (m *MessageReactionCountUpdate) Marshal(noindent ...bool) string {
	if m == nil || m.OriginalUpdate == nil {
		return "null"
	}
	return MarshalWithTypeName(m.OriginalUpdate, noindent...)
}

// BotStopped reports that a user blocked or unblocked the bot.
type BotStopped struct {
	Client         *Client
	OriginalUpdate *UpdateBotStopped
	UserID         int64
	Sender         *UserObj
	Date           int32
	// Stopped is true when the user blocked the bot, false when the user
	// restarted it.
	Stopped bool
	Qts     int32
}

func (b *BotStopped) Marshal(noindent ...bool) string {
	if b == nil || b.OriginalUpdate == nil {
		return "null"
	}
	return MarshalWithTypeName(b.OriginalUpdate, noindent...)
}

// ChatBoostUpdate reports a new or changed boost on a channel the bot
// administers.
type ChatBoostUpdate struct {
	Client         *Client
	OriginalUpdate *UpdateBotChatBoost
	Peer           Peer
	ChatID         int64
	Channel        *Channel
	Boost          *Boost
	Qts            int32
}

// BoosterID returns the user that applied the boost, when not anonymous.
func (b *ChatBoostUpdate) BoosterID() int64 {
	if b == nil || b.Boost == nil {
		return 0
	}
	return b.Boost.UserID
}

func (b *ChatBoostUpdate) Marshal(noindent ...bool) string {
	if b == nil || b.OriginalUpdate == nil {
		return "null"
	}
	return MarshalWithTypeName(b.OriginalUpdate, noindent...)
}

// PurchasedPaidMedia reports that a user bought paid media sent by the bot.
type PurchasedPaidMedia struct {
	Client         *Client
	OriginalUpdate *UpdateBotPurchasedPaidMedia
	UserID         int64
	Sender         *UserObj
	Payload        string
	Qts            int32
}

func (p *PurchasedPaidMedia) Marshal(noindent ...bool) string {
	if p == nil || p.OriginalUpdate == nil {
		return "null"
	}
	return MarshalWithTypeName(p.OriginalUpdate, noindent...)
}

// StarsSubscriptionUpdate reports a change to a Telegram Stars subscription
// sold by the bot.
type StarsSubscriptionUpdate struct {
	Client         *Client
	OriginalUpdate *UpdateBotStarsSubscription
	UserID         int64
	Sender         *UserObj
	Payload        []byte
	Canceled       bool
	PaymentFailed  bool
	Restored       bool
	Qts            int32
}

// PayloadString returns the subscription payload as a string.
func (s *StarsSubscriptionUpdate) PayloadString() string {
	if s == nil {
		return ""
	}
	return string(s.Payload)
}

func (s *StarsSubscriptionUpdate) Marshal(noindent ...bool) string {
	if s == nil || s.OriginalUpdate == nil {
		return "null"
	}
	return MarshalWithTypeName(s.OriginalUpdate, noindent...)
}

// ChatParticipantUpdate reports a membership change in a basic group. Channel
// and supergroup changes are delivered as ParticipantUpdate instead.
type ChatParticipantUpdate struct {
	Client         *Client
	OriginalUpdate *UpdateChatParticipant
	ChatID         int64
	Chat           *ChatObj
	Date           int32
	ActorID        int64
	UserID         int64
	User           *UserObj
	Actor          *UserObj
	Old            ChatParticipant
	New            ChatParticipant
	Invite         ExportedChatInvite
	Qts            int32
}

// Joined reports whether the user just joined the group.
func (p *ChatParticipantUpdate) Joined() bool {
	return p != nil && p.Old == nil && p.New != nil
}

// Left reports whether the user just left or was removed from the group.
func (p *ChatParticipantUpdate) Left() bool {
	return p != nil && p.Old != nil && p.New == nil
}

func (p *ChatParticipantUpdate) Marshal(noindent ...bool) string {
	if p == nil || p.OriginalUpdate == nil {
		return "null"
	}
	return MarshalWithTypeName(p.OriginalUpdate, noindent...)
}

// PollVote reports that a user voted in a public poll.
type PollVote struct {
	Client         *Client
	OriginalUpdate *UpdateMessagePollVote
	PollID         int64
	Peer           Peer
	UserID         int64
	Sender         *UserObj
	Options        [][]byte
	Positions      []int32
	Qts            int32
}

// OptionStrings returns the chosen options as strings.
func (p *PollVote) OptionStrings() []string {
	if p == nil {
		return nil
	}
	out := make([]string, 0, len(p.Options))
	for _, o := range p.Options {
		out = append(out, string(o))
	}
	return out
}

func (p *PollVote) Marshal(noindent ...bool) string {
	if p == nil || p.OriginalUpdate == nil {
		return "null"
	}
	return MarshalWithTypeName(p.OriginalUpdate, noindent...)
}

// PollUpdate reports new results for a poll, including when it is closed.
type PollUpdate struct {
	Client         *Client
	OriginalUpdate *UpdateMessagePoll
	PollID         int64
	Peer           Peer
	ChatID         int64
	MessageID      int32
	Poll           *Poll
	Results        *PollResults
}

// Closed reports whether the poll has been closed.
func (p *PollUpdate) Closed() bool {
	if p == nil || p.Poll == nil {
		return false
	}
	return p.Poll.Closed
}

func (p *PollUpdate) Marshal(noindent ...bool) string {
	if p == nil || p.OriginalUpdate == nil {
		return "null"
	}
	return MarshalWithTypeName(p.OriginalUpdate, noindent...)
}

// GenerationStopped reports that the peer pressed "stop" on a streaming
// message draft that was published with CanStop enabled. It corresponds to
// the Bot API 10.3 MessageGenerationStopped/stopped_message_generation update
// and is carried over MTProto as sendMessageStopDraftAction inside a typing
// update.
//
// RandomID identifies the draft that was stopped and matches the value used
// by the RichDraft that published it, so a bot streaming several drafts can
// tell which one the user interrupted.
type GenerationStopped struct {
	Client         *Client
	OriginalUpdate Update

	RandomID int64
	// Peer is the chat the draft was being streamed to, and ChatID its id.
	Peer   Peer
	ChatID int64
	// Actor is whoever pressed stop. It matches Peer in private chats and may
	// be the group itself when an anonymous admin acts.
	Actor   Peer
	ActorID int64
	// UserID and Sender are only set when the actor is a user.
	UserID   int64
	Sender   *UserObj
	Chat     *ChatObj
	Channel  *Channel
	TopMsgID int32
}

// IsPrivate reports whether the stop happened in a private chat.
func (g *GenerationStopped) IsPrivate() bool {
	if g == nil {
		return false
	}
	_, ok := g.Peer.(*PeerUser)
	return ok
}

// PeerID returns the id of the chat the draft was being streamed to.
func (g *GenerationStopped) PeerID() int64 {
	if g == nil {
		return 0
	}
	return g.ChatID
}

func (g *GenerationStopped) Marshal(noindent ...bool) string {
	if g == nil || g.OriginalUpdate == nil {
		return "null"
	}
	return MarshalWithTypeName(g.OriginalUpdate, noindent...)
}

// ---------------------------- Handler Types ----------------------------

type BusinessMessageHandler func(m *BusinessMessage) error
type BusinessEditHandler func(m *BusinessMessage) error
type BusinessDeleteHandler func(m *BusinessMessageDeleted) error
type BusinessConnectHandler func(m *BusinessConnection) error
type PreCheckoutHandler func(m *PreCheckoutQuery) error
type ShippingHandler func(m *ShippingQuery) error
type MessageReactionHandler func(m *MessageReactionUpdate) error
type MessageReactionCountHandler func(m *MessageReactionCountUpdate) error
type BotStoppedHandler func(m *BotStopped) error
type ChatBoostHandler func(m *ChatBoostUpdate) error
type PurchasedPaidMediaHandler func(m *PurchasedPaidMedia) error
type StarsSubscriptionHandler func(m *StarsSubscriptionUpdate) error
type ChatParticipantHandler func(m *ChatParticipantUpdate) error
type PollVoteHandler func(m *PollVote) error
type PollHandler func(m *PollUpdate) error
type GenerationStoppedHandler func(m *GenerationStopped) error

// botUpdateHandle is the generic handle used by every event above. The older
// handlers each declare their own struct; a single generic type keeps this
// addition small while still satisfying the Handle interface.
type botUpdateHandle[H any] struct {
	baseHandle
	Handler H
}

// ---------------------------- Packers ----------------------------

// senderFor resolves a user without hitting the network for an absent ID.
// Flag-gated fields are frequently zero, and a lookup for peer 0 is both
// meaningless and a wasted round trip.
func senderFor(c *Client, userID int64) *UserObj {
	if userID == 0 {
		return &UserObj{}
	}
	return c.getSender(&PeerUser{UserID: userID})
}

// chatFor resolves a basic group, skipping the lookup for an absent ID.
func chatFor(c *Client, chatID int64) *ChatObj {
	if chatID == 0 {
		return nil
	}
	return c.getChat(&PeerChat{ChatID: chatID})
}

// peerChatFor and peerChannelFor resolve a peer only when it is present and
// carries a non-zero ID, avoiding lookups for absent flag-gated peers.
func peerChatFor(c *Client, p Peer) *ChatObj {
	chat, ok := p.(*PeerChat)
	if !ok || chat == nil || chat.ChatID == 0 {
		return nil
	}
	return c.getChat(chat)
}

func peerChannelFor(c *Client, p Peer) *Channel {
	channel, ok := p.(*PeerChannel)
	if !ok || channel == nil || channel.ChannelID == 0 {
		return nil
	}
	return c.getChannel(channel)
}

func packBusinessMessage(c *Client, connectionID string, msg, replyTo Message, qts int32, raw Update) *BusinessMessage {
	if msg == nil {
		return nil
	}
	b := &BusinessMessage{
		Client:         c,
		OriginalUpdate: raw,
		ConnectionID:   connectionID,
		Message:        packMessage(c, msg),
		Qts:            qts,
	}
	if replyTo != nil {
		b.ReplyTo = packMessage(c, replyTo)
	}
	return b
}

func packBusinessDelete(c *Client, u *UpdateBotDeleteBusinessMessage) *BusinessMessageDeleted {
	if u == nil {
		return nil
	}
	return &BusinessMessageDeleted{
		Client:         c,
		OriginalUpdate: u,
		ConnectionID:   u.ConnectionID,
		Peer:           u.Peer,
		ChatID:         c.GetPeerID(u.Peer),
		Messages:       u.Messages,
		Qts:            u.Qts,
	}
}

func packBusinessConnect(c *Client, u *UpdateBotBusinessConnect) *BusinessConnection {
	if u == nil {
		return nil
	}
	b := &BusinessConnection{
		Client:         c,
		OriginalUpdate: u,
		Connection:     u.Connection,
		Qts:            u.Qts,
	}
	if u.Connection != nil {
		b.ConnectionID = u.Connection.ConnectionID
		b.UserID = u.Connection.UserID
	}
	return b
}

func packPreCheckout(c *Client, u *UpdateBotPrecheckoutQuery) *PreCheckoutQuery {
	if u == nil {
		return nil
	}
	return &PreCheckoutQuery{
		Client:           c,
		OriginalUpdate:   u,
		QueryID:          u.QueryID,
		UserID:           u.UserID,
		Sender:           senderFor(c, u.UserID),
		Payload:          u.Payload,
		Info:             u.Info,
		ShippingOptionID: u.ShippingOptionID,
		Currency:         u.Currency,
		TotalAmount:      u.TotalAmount,
	}
}

func packShipping(c *Client, u *UpdateBotShippingQuery) *ShippingQuery {
	if u == nil {
		return nil
	}
	return &ShippingQuery{
		Client:          c,
		OriginalUpdate:  u,
		QueryID:         u.QueryID,
		UserID:          u.UserID,
		Sender:          senderFor(c, u.UserID),
		Payload:         u.Payload,
		ShippingAddress: u.ShippingAddress,
	}
}

func packMessageReaction(c *Client, u *UpdateBotMessageReaction) *MessageReactionUpdate {
	if u == nil {
		return nil
	}
	m := &MessageReactionUpdate{
		Client:         c,
		OriginalUpdate: u,
		Peer:           u.Peer,
		ChatID:         c.GetPeerID(u.Peer),
		MessageID:      u.MsgID,
		Date:           u.Date,
		Actor:          u.Actor,
		ActorID:        c.GetPeerID(u.Actor),
		Chat:           peerChatFor(c, u.Peer),
		Channel:        peerChannelFor(c, u.Peer),
		OldReactions:   u.OldReactions,
		NewReactions:   u.NewReactions,
	}
	if p, ok := u.Actor.(*PeerUser); ok {
		m.Sender = senderFor(c, p.UserID)
	}
	return m
}

func packMessageReactionCount(c *Client, u *UpdateBotMessageReactions) *MessageReactionCountUpdate {
	if u == nil {
		return nil
	}
	return &MessageReactionCountUpdate{
		Client:         c,
		OriginalUpdate: u,
		Peer:           u.Peer,
		ChatID:         c.GetPeerID(u.Peer),
		MessageID:      u.MsgID,
		Date:           u.Date,
		Chat:           peerChatFor(c, u.Peer),
		Channel:        peerChannelFor(c, u.Peer),
		Reactions:      u.Reactions,
	}
}

func packBotStopped(c *Client, u *UpdateBotStopped) *BotStopped {
	if u == nil {
		return nil
	}
	return &BotStopped{
		Client:         c,
		OriginalUpdate: u,
		UserID:         u.UserID,
		Sender:         senderFor(c, u.UserID),
		Date:           u.Date,
		Stopped:        u.Stopped,
		Qts:            u.Qts,
	}
}

func packChatBoost(c *Client, u *UpdateBotChatBoost) *ChatBoostUpdate {
	if u == nil {
		return nil
	}
	return &ChatBoostUpdate{
		Client:         c,
		OriginalUpdate: u,
		Peer:           u.Peer,
		ChatID:         c.GetPeerID(u.Peer),
		Channel:        peerChannelFor(c, u.Peer),
		Boost:          u.Boost,
		Qts:            u.Qts,
	}
}

func packPurchasedPaidMedia(c *Client, u *UpdateBotPurchasedPaidMedia) *PurchasedPaidMedia {
	if u == nil {
		return nil
	}
	return &PurchasedPaidMedia{
		Client:         c,
		OriginalUpdate: u,
		UserID:         u.UserID,
		Sender:         senderFor(c, u.UserID),
		Payload:        u.Payload,
		Qts:            u.Qts,
	}
}

func packStarsSubscription(c *Client, u *UpdateBotStarsSubscription) *StarsSubscriptionUpdate {
	if u == nil {
		return nil
	}
	return &StarsSubscriptionUpdate{
		Client:         c,
		OriginalUpdate: u,
		UserID:         u.UserID,
		Sender:         senderFor(c, u.UserID),
		Payload:        u.Payload,
		Canceled:       u.Canceled,
		PaymentFailed:  u.PaymentFailed,
		Restored:       u.Restored,
		Qts:            u.Qts,
	}
}

func packChatParticipant(c *Client, u *UpdateChatParticipant) *ChatParticipantUpdate {
	if u == nil {
		return nil
	}
	return &ChatParticipantUpdate{
		Client:         c,
		OriginalUpdate: u,
		ChatID:         u.ChatID,
		Chat:           chatFor(c, u.ChatID),
		Date:           u.Date,
		ActorID:        u.ActorID,
		UserID:         u.UserID,
		User:           senderFor(c, u.UserID),
		Actor:          senderFor(c, u.ActorID),
		Old:            u.PrevParticipant,
		New:            u.NewParticipant,
		Invite:         u.Invite,
		Qts:            u.Qts,
	}
}

func packPollVote(c *Client, u *UpdateMessagePollVote) *PollVote {
	if u == nil {
		return nil
	}
	v := &PollVote{
		Client:         c,
		OriginalUpdate: u,
		PollID:         u.PollID,
		Peer:           u.Peer,
		UserID:         c.GetPeerID(u.Peer),
		Options:        u.Options,
		Positions:      u.Positions,
		Qts:            u.Qts,
	}
	if p, ok := u.Peer.(*PeerUser); ok {
		v.Sender = senderFor(c, p.UserID)
	}
	return v
}

func packPollUpdate(c *Client, u *UpdateMessagePoll) *PollUpdate {
	if u == nil {
		return nil
	}
	return &PollUpdate{
		Client:         c,
		OriginalUpdate: u,
		PollID:         u.PollID,
		Peer:           u.Peer,
		ChatID:         c.GetPeerID(u.Peer),
		MessageID:      u.MsgID,
		Poll:           u.Poll,
		Results:        u.Results,
	}
}

// isStopDraftAction reports whether a typing action is a generation stop.
// Typing updates are by far the most frequent updates on the wire, so the
// dispatcher checks this inline before spawning a goroutine for one.
func isStopDraftAction(action SendMessageAction) bool {
	stop, ok := action.(*SendMessageStopDraftAction)
	return ok && stop != nil
}

// packGenerationStopped builds a GenerationStopped from any of the three
// typing updates. Telegram delivers the stop signal as a sendMessageAction, so
// other action types yield nil and are ignored.
//
// peer is the chat the draft was streamed to and actor is whoever pressed
// stop; they coincide in private chats.
func packGenerationStopped(c *Client, action SendMessageAction, peer, actor Peer, topMsgID int32, raw Update) *GenerationStopped {
	stop, ok := action.(*SendMessageStopDraftAction)
	if !ok || stop == nil {
		return nil
	}

	g := &GenerationStopped{
		Client:         c,
		OriginalUpdate: raw,
		RandomID:       stop.RandomID,
		Peer:           peer,
		ChatID:         c.GetPeerID(peer),
		Actor:          actor,
		ActorID:        c.GetPeerID(actor),
		Chat:           peerChatFor(c, peer),
		Channel:        peerChannelFor(c, peer),
		TopMsgID:       topMsgID,
	}
	if p, ok := actor.(*PeerUser); ok {
		g.UserID = p.UserID
		g.Sender = senderFor(c, p.UserID)
	}

	return g
}

// ---------------------------- Dispatch ----------------------------

// dispatchBotUpdate runs every registered handler for a packed event, mirroring
// the group/priority and ErrEndGroup semantics of the pre-existing handlers.
// E is constrained to a pointer so a nil packed event can be skipped.
func dispatchBotUpdate[H any, T any](c *Client, handles map[int][]*botUpdateHandle[H], packed *T, invoke func(H, *T) error, middlewares []func(H) H, logLabel string) {
	if packed == nil || len(handles) == 0 {
		return
	}

	c.dispatcher.RLock()
	snapshot := make(map[int][]*botUpdateHandle[H], len(handles))
	maps.Copy(snapshot, handles)
	c.dispatcher.RUnlock()

	for group, handlers := range snapshot {
		for _, handler := range handlers {
			if handler == nil {
				continue
			}
			handle := func(h *botUpdateHandle[H]) error {
				defer c.NewRecovery()()
				hf := applyChain(h.Handler, middlewares)
				return invoke(hf, packed)
			}

			if group == DefaultGroup {
				go func() {
					if err := handle(handler); err != nil && !errors.Is(err, ErrEndGroup) {
						c.Log.WithError(err).Error(logLabel)
					}
				}()
			} else {
				if err := handle(handler); err != nil && errors.Is(err, ErrEndGroup) {
					break
				}
			}
		}
	}
}

func (c *Client) handleBusinessMessageUpdate(u *UpdateBotNewBusinessMessage) {
	if u == nil {
		return
	}
	packed := packBusinessMessage(c, u.ConnectionID, u.Message, u.ReplyToMessage, u.Qts, u)
	dispatchBotUpdate(c, c.dispatcher.businessMessageHandles, packed,
		func(h BusinessMessageHandler, m *BusinessMessage) error { return h(m) },
		c.dispatcher.middlewareManager.businessMessages(), "[BusinessMessageHandler]")
}

func (c *Client) handleBusinessEditUpdate(u *UpdateBotEditBusinessMessage) {
	if u == nil {
		return
	}
	packed := packBusinessMessage(c, u.ConnectionID, u.Message, u.ReplyToMessage, u.Qts, u)
	dispatchBotUpdate(c, c.dispatcher.businessEditHandles, packed,
		func(h BusinessEditHandler, m *BusinessMessage) error { return h(m) },
		c.dispatcher.middlewareManager.businessEdits(), "[BusinessEditHandler]")
}

func (c *Client) handleBusinessDeleteUpdate(u *UpdateBotDeleteBusinessMessage) {
	dispatchBotUpdate(c, c.dispatcher.businessDeleteHandles, packBusinessDelete(c, u),
		func(h BusinessDeleteHandler, m *BusinessMessageDeleted) error { return h(m) },
		c.dispatcher.middlewareManager.businessDeletes(), "[BusinessDeleteHandler]")
}

func (c *Client) handleBusinessConnectUpdate(u *UpdateBotBusinessConnect) {
	dispatchBotUpdate(c, c.dispatcher.businessConnectHandles, packBusinessConnect(c, u),
		func(h BusinessConnectHandler, m *BusinessConnection) error { return h(m) },
		c.dispatcher.middlewareManager.businessConnects(), "[BusinessConnectHandler]")
}

func (c *Client) handlePreCheckoutUpdate(u *UpdateBotPrecheckoutQuery) {
	dispatchBotUpdate(c, c.dispatcher.preCheckoutHandles, packPreCheckout(c, u),
		func(h PreCheckoutHandler, m *PreCheckoutQuery) error { return h(m) },
		c.dispatcher.middlewareManager.preCheckouts(), "[PreCheckoutHandler]")
}

func (c *Client) handleShippingUpdate(u *UpdateBotShippingQuery) {
	dispatchBotUpdate(c, c.dispatcher.shippingHandles, packShipping(c, u),
		func(h ShippingHandler, m *ShippingQuery) error { return h(m) },
		c.dispatcher.middlewareManager.shippings(), "[ShippingHandler]")
}

func (c *Client) handleMessageReactionUpdate(u *UpdateBotMessageReaction) {
	dispatchBotUpdate(c, c.dispatcher.reactionHandles, packMessageReaction(c, u),
		func(h MessageReactionHandler, m *MessageReactionUpdate) error { return h(m) },
		c.dispatcher.middlewareManager.reactions(), "[MessageReactionHandler]")
}

func (c *Client) handleMessageReactionCountUpdate(u *UpdateBotMessageReactions) {
	dispatchBotUpdate(c, c.dispatcher.reactionCountHandles, packMessageReactionCount(c, u),
		func(h MessageReactionCountHandler, m *MessageReactionCountUpdate) error { return h(m) },
		c.dispatcher.middlewareManager.reactionCounts(), "[MessageReactionCountHandler]")
}

func (c *Client) handleBotStoppedUpdate(u *UpdateBotStopped) {
	dispatchBotUpdate(c, c.dispatcher.botStoppedHandles, packBotStopped(c, u),
		func(h BotStoppedHandler, m *BotStopped) error { return h(m) },
		c.dispatcher.middlewareManager.botStoppeds(), "[BotStoppedHandler]")
}

func (c *Client) handleChatBoostUpdate(u *UpdateBotChatBoost) {
	dispatchBotUpdate(c, c.dispatcher.chatBoostHandles, packChatBoost(c, u),
		func(h ChatBoostHandler, m *ChatBoostUpdate) error { return h(m) },
		c.dispatcher.middlewareManager.chatBoosts(), "[ChatBoostHandler]")
}

func (c *Client) handlePurchasedPaidMediaUpdate(u *UpdateBotPurchasedPaidMedia) {
	dispatchBotUpdate(c, c.dispatcher.paidMediaHandles, packPurchasedPaidMedia(c, u),
		func(h PurchasedPaidMediaHandler, m *PurchasedPaidMedia) error { return h(m) },
		c.dispatcher.middlewareManager.paidMedias(), "[PurchasedPaidMediaHandler]")
}

func (c *Client) handleStarsSubscriptionUpdate(u *UpdateBotStarsSubscription) {
	dispatchBotUpdate(c, c.dispatcher.starsSubHandles, packStarsSubscription(c, u),
		func(h StarsSubscriptionHandler, m *StarsSubscriptionUpdate) error { return h(m) },
		c.dispatcher.middlewareManager.starsSubs(), "[StarsSubscriptionHandler]")
}

func (c *Client) handleChatParticipantUpdate(u *UpdateChatParticipant) {
	dispatchBotUpdate(c, c.dispatcher.chatParticipantHandles, packChatParticipant(c, u),
		func(h ChatParticipantHandler, m *ChatParticipantUpdate) error { return h(m) },
		c.dispatcher.middlewareManager.chatParticipants(), "[ChatParticipantHandler]")
}

func (c *Client) handlePollVoteUpdate(u *UpdateMessagePollVote) {
	dispatchBotUpdate(c, c.dispatcher.pollVoteHandles, packPollVote(c, u),
		func(h PollVoteHandler, m *PollVote) error { return h(m) },
		c.dispatcher.middlewareManager.pollVotes(), "[PollVoteHandler]")
}

func (c *Client) handlePollUpdate(u *UpdateMessagePoll) {
	dispatchBotUpdate(c, c.dispatcher.pollHandles, packPollUpdate(c, u),
		func(h PollHandler, m *PollUpdate) error { return h(m) },
		c.dispatcher.middlewareManager.polls(), "[PollHandler]")
}

// dispatchGenerationStopped notifies any live RichDraft first so that
// draft.Stopped() flips as soon as the update arrives instead of only after
// the next draft action is rejected, then runs the registered handlers.
func (c *Client) dispatchGenerationStopped(g *GenerationStopped) {
	if g == nil {
		return
	}
	c.markDraftStopped(g.RandomID)
	dispatchBotUpdate(c, c.dispatcher.generationStoppedHandles, g,
		func(h GenerationStoppedHandler, m *GenerationStopped) error { return h(m) },
		c.dispatcher.middlewareManager.generationStoppeds(), "[GenerationStoppedHandler]")
}

// handleUserTypingUpdate covers private chats, where the typing peer is also
// the chat the draft was streamed to.
func (c *Client) handleUserTypingUpdate(u *UpdateUserTyping) {
	if u == nil {
		return
	}
	peer := &PeerUser{UserID: u.UserID}
	c.dispatchGenerationStopped(packGenerationStopped(c, u.Action, peer, peer, u.TopMsgID, u))
}

func (c *Client) handleChatTypingUpdate(u *UpdateChatUserTyping) {
	if u == nil {
		return
	}
	c.dispatchGenerationStopped(packGenerationStopped(c, u.Action, &PeerChat{ChatID: u.ChatID}, u.FromID, 0, u))
}

func (c *Client) handleChannelTypingUpdate(u *UpdateChannelUserTyping) {
	if u == nil {
		return
	}
	c.dispatchGenerationStopped(packGenerationStopped(c, u.Action, &PeerChannel{ChannelID: u.ChannelID}, u.FromID, u.TopMsgID, u))
}

// ---------------------------- Registration ----------------------------

func addBotUpdateHandle[H any](c *Client, handles map[int][]*botUpdateHandle[H], handler H) Handle {
	c.dispatcher.Lock()
	defer c.dispatcher.Unlock()
	handleID := nextHandleID()
	h := &botUpdateHandle[H]{
		Handler:    handler,
		baseHandle: baseHandle{id: handleID, Group: DefaultGroup},
	}
	h.onGroupChanged = makeGroupChangeCallback(handles, h, handleID, &c.dispatcher.RWMutex)
	h.onPriorityChanged = makePriorityChangeCallback(handles, h, handleID, h.GetGroup, h.GetPriority, &c.dispatcher.RWMutex)
	return addHandleToMap(handles, h)
}

// OnBusinessMessage registers a handler for messages received through a
// Telegram Business connection.
func (c *Client) OnBusinessMessage(handler BusinessMessageHandler) Handle {
	return addBotUpdateHandle(c, c.dispatcher.businessMessageHandles, handler)
}

// OnBusinessEdit registers a handler for edits in connected business chats.
func (c *Client) OnBusinessEdit(handler BusinessEditHandler) Handle {
	return addBotUpdateHandle(c, c.dispatcher.businessEditHandles, handler)
}

// OnBusinessDelete registers a handler for deletions in connected business chats.
func (c *Client) OnBusinessDelete(handler BusinessDeleteHandler) Handle {
	return addBotUpdateHandle(c, c.dispatcher.businessDeleteHandles, handler)
}

// OnBusinessConnect registers a handler for business connection changes.
func (c *Client) OnBusinessConnect(handler BusinessConnectHandler) Handle {
	return addBotUpdateHandle(c, c.dispatcher.businessConnectHandles, handler)
}

// OnPreCheckout registers a handler for pre-checkout queries. The handler must
// answer within 10 seconds or the payment is cancelled by Telegram.
func (c *Client) OnPreCheckout(handler PreCheckoutHandler) Handle {
	return addBotUpdateHandle(c, c.dispatcher.preCheckoutHandles, handler)
}

// OnShipping registers a handler for shipping queries.
func (c *Client) OnShipping(handler ShippingHandler) Handle {
	return addBotUpdateHandle(c, c.dispatcher.shippingHandles, handler)
}

// OnReaction registers a handler for per-user message reaction changes.
func (c *Client) OnReaction(handler MessageReactionHandler) Handle {
	return addBotUpdateHandle(c, c.dispatcher.reactionHandles, handler)
}

// OnReactionCount registers a handler for anonymous reaction counters in large
// groups and channels.
func (c *Client) OnReactionCount(handler MessageReactionCountHandler) Handle {
	return addBotUpdateHandle(c, c.dispatcher.reactionCountHandles, handler)
}

// OnBotStopped registers a handler for users blocking or restarting the bot.
func (c *Client) OnBotStopped(handler BotStoppedHandler) Handle {
	return addBotUpdateHandle(c, c.dispatcher.botStoppedHandles, handler)
}

// OnChatBoost registers a handler for channel boost changes.
func (c *Client) OnChatBoost(handler ChatBoostHandler) Handle {
	return addBotUpdateHandle(c, c.dispatcher.chatBoostHandles, handler)
}

// OnPaidMediaPurchase registers a handler for paid media purchases.
func (c *Client) OnPaidMediaPurchase(handler PurchasedPaidMediaHandler) Handle {
	return addBotUpdateHandle(c, c.dispatcher.paidMediaHandles, handler)
}

// OnStarsSubscription registers a handler for Telegram Stars subscription
// changes.
func (c *Client) OnStarsSubscription(handler StarsSubscriptionHandler) Handle {
	return addBotUpdateHandle(c, c.dispatcher.starsSubHandles, handler)
}

// OnChatParticipant registers a handler for basic-group membership changes.
// Supergroup and channel changes are delivered to OnParticipant.
func (c *Client) OnChatParticipant(handler ChatParticipantHandler) Handle {
	return addBotUpdateHandle(c, c.dispatcher.chatParticipantHandles, handler)
}

// OnPollVote registers a handler for votes in public polls.
func (c *Client) OnPollVote(handler PollVoteHandler) Handle {
	return addBotUpdateHandle(c, c.dispatcher.pollVoteHandles, handler)
}

// OnPoll registers a handler for poll result updates.
func (c *Client) OnPoll(handler PollHandler) Handle {
	return addBotUpdateHandle(c, c.dispatcher.pollHandles, handler)
}

// OnGenerationStopped registers a handler for the peer stopping a streaming
// message draft that was published with CanStop enabled (Bot API 10.3
// stopped_message_generation).
//
// A RichDraft created by this client also has its Stopped() flag set before
// the handler runs, so streaming loops can simply poll draft.Stopped().
func (c *Client) OnGenerationStopped(handler GenerationStoppedHandler) Handle {
	return addBotUpdateHandle(c, c.dispatcher.generationStoppedHandles, handler)
}
