// +build linux

package notifier

import (
	"github.com/esiqveland/notify"
	"github.com/godbus/dbus"
	"path/filepath"
)

type Notifier struct {
	id       uint32
	notifier notify.Notifier
	icon     string
}

// Init connects notifier to DBus session.
func (n *Notifier) Init(icon string) error {
	if path, err := filepath.Abs(icon); err != nil {
		return err
	} else {
		n.icon = path
	}
	if conn, err := dbus.SessionBus(); err != nil {
		return err
	} else {
		if notifier, err := notify.New(conn); err != nil {
			return err
		} else {
			n.notifier = notifier
			return nil
		}
	}
}

// AddNotifyIcon adds a notification balloon with passed tooltip, title and description.
func (n *Notifier) AddNotifyIcon(tip, title, info string) error {
	return n.Update(tip, title, info)
}

// Update updates existing notification balloon with passed tooltip, title and description.
func (n *Notifier) Update(tip, title, info string) error {
	id, err := n.notifier.SendNotification(notify.Notification{
		AppName:       tip,
		ReplacesID:    n.id,
		AppIcon:       n.icon,
		Summary:       title,
		Body:          info,
		Hints:         map[string]dbus.Variant{},
		ExpireTimeout: int32(5000),
	})
	if err != nil {
		return err
	} else {
		n.id = id
		return nil
	}
}

// Close closes notification balloon.
func (n *Notifier) Close() {
	if _, err := n.notifier.CloseNotification(int(n.id)); err != nil {
		panic(err)
	}

	if err := n.notifier.Close(); err != nil {
		panic(err)
	}
}

// NewNotifier creates notification popup wrapper with custom icon.
func NewNotifier(icon string) (*Notifier, error) {
	n := new(Notifier)
	if err := n.Init(icon); err != nil {
		return nil, err
	} else {
		return n, nil
	}
}
