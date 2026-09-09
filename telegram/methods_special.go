// Copyright (c) 2025 @AmarnathCJD

package telegram

import (
	"fmt"

	"github.com/amarnathcjd/gogram/internal/encoding/tl"
)

type InitConnectionParams struct {
	ApiID          int32             // Application identifier (see. App configuration)
	DeviceModel    string            // Device model
	SystemVersion  string            // Operation system version
	AppVersion     string            // Application version
	SystemLangCode string            // Code for the language used on the device's OS, ISO 639-1 standard
	LangPack       string            // Language pack to use
	LangCode       string            // Code for the language used on the client, ISO 639-1 standard
	Proxy          *InputClientProxy `tl:"flag:0"` // Info about an MTProto proxy
	Params         JsonValue         `tl:"flag:1"` // Additional initConnection parameters. For now, only the tz_offset field is supported, for specifying timezone offset in seconds.
	Query          tl.Object         // The query itself
}

func (*InitConnectionParams) CRC() uint32 {
	return 0xc1cd5ea9
}

func (*InitConnectionParams) FlagIndex() int {
	return 0
}

func (c *Client) InitConnection(params *InitConnectionParams) (tl.Object, error) {
	data, err := c.MakeRequest(params)
	if err != nil {
		return nil, fmt.Errorf("sending InitConnection: %w", err)
	}

	return data.(tl.Object), nil
}

type InvokeWithLayerParams struct {
	Layer int32
	Query tl.Object
}

func (*InvokeWithLayerParams) CRC() uint32 {
	return 0xda9b0d0d
}

func (c *Client) InvokeWithLayer(layer int, query tl.Object) (tl.Object, error) {
	data, err := c.MakeRequest(&InvokeWithLayerParams{
		Layer: int32(layer),
		Query: query,
	})
	if err != nil {
		return nil, fmt.Errorf("sending InvokeWithLayer: %w", err)
	}

	return data.(tl.Object), nil
}

type InvokeWithoutUpdatesParams struct {
	Query tl.Object
}

func (*InvokeWithoutUpdatesParams) CRC() uint32 {
	return 0xbf9459b7
}

func (c *Client) InvokeWithoutUpdates(query tl.Object) (tl.Object, error) {
	data, err := c.MakeRequest(&InvokeWithoutUpdatesParams{
		Query: query,
	})
	if err != nil {
		return nil, fmt.Errorf("sending InvokeWithoutUpdates: %w", err)
	}
	return data.(tl.Object), nil
}

type InvokeWithMessagesRangeParams struct {
	Range MessageRange
	Query tl.Object
}

func (*InvokeWithMessagesRangeParams) CRC() uint32 {
	return 0x365275f2
}

func (c *Client) InvokeWithMessagesRange(r MessageRange, query tl.Object) (tl.Object, error) {
	data, err := c.MakeRequest(&InvokeWithMessagesRangeParams{
		Range: r,
		Query: query,
	})
	if err != nil {
		return nil, fmt.Errorf("sending InvokeWithMessagesRange: %w", err)
	}
	return data.(tl.Object), nil
}

type InvokeWithTakeoutParams struct {
	TakeoutID int64
	Query     tl.Object
}

func (*InvokeWithTakeoutParams) CRC() uint32 {
	return 0xaca9fd2e
}

func (c *Client) InvokeWithTakeout(takeoutID int64, query tl.Object) (tl.Object, error) {
	data, err := c.MakeRequest(&InvokeWithTakeoutParams{
		TakeoutID: takeoutID,
		Query:     query,
	})
	if err != nil {
		return nil, fmt.Errorf("sending InvokeWithTakeout: %w", err)
	}

	return data.(tl.Object), nil
}

// invokeAfterMsg#cb9f372d {X:Type} msg_id:long query:!X = X;
type InvokeAfterMsgParams struct {
	MsgID int64     // Message identifier on which a current query depends
	Query tl.Object // The query itself
}

func (*InvokeAfterMsgParams) CRC() uint32 {
	return 0xcb9f372d
}

// InvokeAfterMsg invokes query only if the specified message is processed first.
func (c *Client) InvokeAfterMsg(msgID int64, query tl.Object) (tl.Object, error) {
	data, err := c.MakeRequest(&InvokeAfterMsgParams{
		MsgID: msgID,
		Query: query,
	})
	if err != nil {
		return nil, fmt.Errorf("sending InvokeAfterMsg: %w", err)
	}

	return data.(tl.Object), nil
}

// invokeAfterMsgs#3dc4b4f0 {X:Type} msg_ids:Vector<long> query:!X = X;
type InvokeAfterMsgsParams struct {
	MsgIDs []int64   // Messages on which a current query depends
	Query  tl.Object // The query itself
}

func (*InvokeAfterMsgsParams) CRC() uint32 {
	return 0x3dc4b4f0
}

// InvokeAfterMsgs invokes query only if all the specified messages are processed first.
func (c *Client) InvokeAfterMsgs(msgIDs []int64, query tl.Object) (tl.Object, error) {
	data, err := c.MakeRequest(&InvokeAfterMsgsParams{
		MsgIDs: msgIDs,
		Query:  query,
	})
	if err != nil {
		return nil, fmt.Errorf("sending InvokeAfterMsgs: %w", err)
	}

	return data.(tl.Object), nil
}

// invokeWithBusinessConnection#dd289f8e {X:Type} connection_id:string query:!X = X;
type InvokeWithBusinessConnectionParams struct {
	ConnectionID string    // Business connection identifier
	Query        tl.Object // The query itself
}

func (*InvokeWithBusinessConnectionParams) CRC() uint32 {
	return 0xdd289f8e
}

// InvokeWithBusinessConnection invokes query on behalf of a user connected to
// the current bot through a business connection.
func (c *Client) InvokeWithBusinessConnection(connectionID string, query tl.Object) (tl.Object, error) {
	data, err := c.MakeRequest(&InvokeWithBusinessConnectionParams{
		ConnectionID: connectionID,
		Query:        query,
	})
	if err != nil {
		return nil, fmt.Errorf("sending InvokeWithBusinessConnection: %w", err)
	}

	return data.(tl.Object), nil
}

// invokeWithGooglePlayIntegrity#1df92984 {X:Type} nonce:string token:string query:!X = X;
type InvokeWithGooglePlayIntegrityParams struct {
	Nonce string    // Nonce obtained from the integrity check request
	Token string    // Token returned by the Google Play Integrity API
	Query tl.Object // The query itself
}

func (*InvokeWithGooglePlayIntegrityParams) CRC() uint32 {
	return 0x1df92984
}

// InvokeWithGooglePlayIntegrity invokes query, passing a Google Play Integrity
// verdict to the server (Android clients only).
func (c *Client) InvokeWithGooglePlayIntegrity(nonce, token string, query tl.Object) (tl.Object, error) {
	data, err := c.MakeRequest(&InvokeWithGooglePlayIntegrityParams{
		Nonce: nonce,
		Token: token,
		Query: query,
	})
	if err != nil {
		return nil, fmt.Errorf("sending InvokeWithGooglePlayIntegrity: %w", err)
	}

	return data.(tl.Object), nil
}

// invokeWithApnsSecret#0dae54f8 {X:Type} nonce:string secret:string query:!X = X;
type InvokeWithApnsSecretParams struct {
	Nonce  string    // Nonce obtained from the integrity check request
	Secret string    // Secret returned by the APNS integrity check
	Query  tl.Object // The query itself
}

func (*InvokeWithApnsSecretParams) CRC() uint32 {
	return 0x0dae54f8
}

// InvokeWithApnsSecret invokes query, passing an APNS integrity secret to the
// server (Apple clients only).
func (c *Client) InvokeWithApnsSecret(nonce, secret string, query tl.Object) (tl.Object, error) {
	data, err := c.MakeRequest(&InvokeWithApnsSecretParams{
		Nonce:  nonce,
		Secret: secret,
		Query:  query,
	})
	if err != nil {
		return nil, fmt.Errorf("sending InvokeWithApnsSecret: %w", err)
	}

	return data.(tl.Object), nil
}

// invokeWithReCaptcha#adbb0f94 {X:Type} token:string query:!X = X;
type InvokeWithReCaptchaParams struct {
	Token string    // Token returned by the reCAPTCHA challenge
	Query tl.Object // The query itself
}

func (*InvokeWithReCaptchaParams) CRC() uint32 {
	return 0xadbb0f94
}

// InvokeWithReCaptcha invokes query, passing a solved reCAPTCHA token to the server.
func (c *Client) InvokeWithReCaptcha(token string, query tl.Object) (tl.Object, error) {
	data, err := c.MakeRequest(&InvokeWithReCaptchaParams{
		Token: token,
		Query: query,
	})
	if err != nil {
		return nil, fmt.Errorf("sending InvokeWithReCaptcha: %w", err)
	}

	return data.(tl.Object), nil
}
