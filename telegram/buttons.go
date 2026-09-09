package telegram

import (
	"bytes"
	"errors"
	"strings"
)

type ButtonBuilder struct{}

var Button = ButtonBuilder{}

type KeyboardBuilder struct {
	rows       []*KeyboardInlineButtonRow
	forceReply bool
}
type ReplyKeyboardBuilder struct{ rows []*KeyboardButtonRow }
type BuildReplyOptions struct {
	ResizeKeyboard, OneTime, Selective, Persistent bool
	ForceReply                                     bool
	Placeholder                                    string
}

func NewKeyboard() *KeyboardBuilder           { return &KeyboardBuilder{} }
func NewReplyKeyboard() *ReplyKeyboardBuilder { return &ReplyKeyboardBuilder{} }
func (k *KeyboardBuilder) AddRow(b ...KeyboardInlineButton) *KeyboardBuilder {
	p := make([]*KeyboardInlineButton, len(b))
	for i := range b {
		p[i] = &b[i]
	}
	k.rows = append(k.rows, &KeyboardInlineButtonRow{Buttons: p})
	return k
}
func (k *KeyboardBuilder) Add(b KeyboardInlineButton) *KeyboardBuilder {
	if len(k.rows) == 0 {
		return k.AddRow(b)
	}
	k.rows[len(k.rows)-1].Buttons = append(k.rows[len(k.rows)-1].Buttons, &b)
	return k
}
func (k *KeyboardBuilder) NewGrid(_, n int, b ...KeyboardInlineButton) *KeyboardBuilder {
	return k.grid(n, b...)
}
func (k *KeyboardBuilder) NewColumn(n int, b ...KeyboardInlineButton) *KeyboardBuilder {
	return k.grid(n, b...)
}
func (k *KeyboardBuilder) NewRow(n int, b ...KeyboardInlineButton) *KeyboardBuilder {
	return k.grid(n, b...)
}
func (k *KeyboardBuilder) grid(n int, b ...KeyboardInlineButton) *KeyboardBuilder {
	if n <= 0 {
		return k
	}
	for i := 0; i < len(b); i += n {
		e := i + n
		if e > len(b) {
			e = len(b)
		}
		k.AddRow(b[i:e]...)
	}
	return k
}

// ForceReply forces the user's client to open a reply interface for the
// message this keyboard is attached to.
func (k *KeyboardBuilder) ForceReply() *KeyboardBuilder {
	k.forceReply = true
	return k
}

func (k *KeyboardBuilder) Build() *ReplyInlineMarkup {
	return &ReplyInlineMarkup{Rows: k.rows, ForceReply: k.forceReply}
}
func (k *ReplyKeyboardBuilder) AddRow(b ...KeyboardButton) *ReplyKeyboardBuilder {
	p := make([]*KeyboardButton, len(b))
	for i := range b {
		p[i] = &b[i]
	}
	k.rows = append(k.rows, &KeyboardButtonRow{Buttons: p})
	return k
}
func (k *ReplyKeyboardBuilder) Build(o ...BuildReplyOptions) *ReplyKeyboardMarkup {
	v := getVariadic(o, BuildReplyOptions{})
	return &ReplyKeyboardMarkup{Resize: v.ResizeKeyboard, SingleUse: v.OneTime, Selective: v.Selective, Persistent: v.Persistent, ForceReply: v.ForceReply, Placeholder: v.Placeholder, Rows: k.rows}
}

func (ButtonBuilder) Data(t, d string) KeyboardInlineButton {
	return KeyboardInlineButton{Text: t, Type: &InlineButtonTypeCallback{Data: []byte(d)}}
}
func (ButtonBuilder) URL(t, u string) KeyboardInlineButton {
	return KeyboardInlineButton{Text: t, Type: &InlineButtonTypeURL{URL: u}}
}
func (ButtonBuilder) SwitchInline(t string, s bool, q string) KeyboardInlineButton {
	return KeyboardInlineButton{Text: t, Type: &InlineButtonTypeSwitchInline{SamePeer: s, Query: q}}
}
func (ButtonBuilder) WebView(t, u string) KeyboardInlineButton {
	return KeyboardInlineButton{Text: t, Type: &InlineButtonTypeWebView{URL: u}}
}
func (ButtonBuilder) Game(t string) KeyboardInlineButton {
	return KeyboardInlineButton{Text: t, Type: &InlineButtonTypeGame{}}
}
func (ButtonBuilder) Buy(t string) KeyboardInlineButton {
	return KeyboardInlineButton{Text: t, Type: &InlineButtonTypeBuy{}}
}
func (ButtonBuilder) Copy(t, c string) KeyboardInlineButton {
	return KeyboardInlineButton{Text: t, Type: &InlineButtonTypeCopy{CopyText: c}}
}

// Disabled returns a greyed-out, non-interactive button. Tapping it does
// nothing, which is useful for placeholders and layout spacing.
func (ButtonBuilder) Disabled(t string) KeyboardInlineButton {
	return KeyboardInlineButton{Text: t, Type: &InlineButtonTypeDisabled{}}
}

// ButtonStyle describes the optional appearance of a keyboard button.
// The three background colors are mutually exclusive; if more than one is
// set, Telegram applies the first in the order primary, danger, success.
type ButtonStyle struct {
	Primary bool  // Accent-colored background
	Danger  bool  // Destructive (red) background
	Success bool  // Affirmative (green) background
	Icon    int64 // Custom emoji ID shown on the button
}

func (s *ButtonStyle) inline() *KeyboardButtonStyle {
	if s == nil {
		return nil
	}
	return &KeyboardButtonStyle{
		BgPrimary: s.Primary,
		BgDanger:  s.Danger,
		BgSuccess: s.Success,
		Icon:      s.Icon,
	}
}

// Styled applies a style to an inline button, returning it for chaining:
//
//	telegram.Button.Styled(telegram.Button.Data("Delete", "rm"), &telegram.ButtonStyle{Danger: true})
func (ButtonBuilder) Styled(b KeyboardInlineButton, s *ButtonStyle) KeyboardInlineButton {
	b.Style = s.inline()
	return b
}

// StyledText returns a reply-keyboard text button with the given style.
func (ButtonBuilder) StyledText(t string, s *ButtonStyle) KeyboardButton {
	return KeyboardButton{Text: t, Type: &ButtonTypeDefault{}, Style: s.inline()}
}
func (ButtonBuilder) Text(t string) KeyboardButton {
	return KeyboardButton{Text: t, Type: &ButtonTypeDefault{}}
}
func (ButtonBuilder) RequestLocation(t string) KeyboardButton {
	return KeyboardButton{Text: t, Type: &ButtonTypeRequestGeoLocation{}}
}
func (ButtonBuilder) RequestPhone(t string) KeyboardButton {
	return KeyboardButton{Text: t, Type: &ButtonTypeRequestPhone{}}
}
func (ButtonBuilder) RequestPoll(t string, q bool) KeyboardButton {
	return KeyboardButton{Text: t, Type: &ButtonTypeRequestPoll{Quiz: q}}
}
func (ButtonBuilder) Clear() *ReplyKeyboardHide { return &ReplyKeyboardHide{} }

func (m *NewMessage) Click(o ...any) (*MessagesBotCallbackAnswer, error) {
	if m.ReplyMarkup() == nil {
		return nil, errors.New("replyMarkup: message has no buttons")
	}
	mk, ok := (*m.ReplyMarkup()).(*ReplyInlineMarkup)
	if !ok {
		return nil, errors.New("replyMarkup: not inline markup")
	}
	for x, r := range mk.Rows {
		for y, b := range r.Buttons {
			match := len(o) == 0 && x == 0 && y == 0
			if len(o) > 0 {
				switch v := o[0].(type) {
				case string:
					match = strings.EqualFold(b.Text, v)
				case []byte:
					t, z := b.Type.(*InlineButtonTypeCallback)
					match = z && bytes.Equal(t.Data, v)
				case []int:
					match = len(v) == 2 && v[0] == x && v[1] == y
				}
			}
			if match {
				if t, z := b.Type.(*InlineButtonTypeCallback); z {
					return m.Client.MessagesGetBotCallbackAnswer(&MessagesGetBotCallbackAnswerParams{Peer: m.Peer, MsgID: m.ID, Data: t.Data})
				}
			}
		}
	}
	return nil, errors.New("replyMarkup: callback button not found")
}
