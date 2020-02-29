// +build windows

package notifier

import (
	"github.com/TheTitanrain/w32"
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

const (
	NimAdd    = 0x00000000
	NimModify = 0x00000001
	NimDelete = 0x00000002

	NifMessage = 0x00000001
	NifIcon    = 0x00000002
	NifTip     = 0x00000004
	NifInfo    = 0x00000010

	ImageIcon      = 1
	LrDefaultSize  = 0x00000040
	LrLoadFromFile = 0x00000010
)

var (
	shell32             = syscall.MustLoadDLL("shell32.dll")
	procShellNotifyIcon = shell32.MustFindProc("Shell_NotifyIconW")
	user32              = windows.NewLazySystemDLL("user32.dll")
	procLoadImageW      = user32.NewProc("LoadImageW")
)

// shellNotifyIcon is a wrapper for Shell_NotifyIconW @ shell32.dll. It passes error from windows api call.
func shellNotifyIcon(dwMessage uint32, lpData *NOTIFYICONDATA) error {
	ptr, _, err := procShellNotifyIcon.Call(uintptr(dwMessage), uintptr(unsafe.Pointer(lpData)))
	if ptr == 0 {
		return err
	}
	return nil
}

// loadImage is a wrapper for LoadImageW @ user32.dll. It returns a handle for loaded image or windows api error.
func loadImage(hInst w32.HINSTANCE, name *uint16, type_ uint32, cx, cy int32, fuLoad uint32) (w32.HICON, error) {
	hicon, _, err := procLoadImageW.Call(uintptr(hInst), uintptr(unsafe.Pointer(name)), uintptr(type_),
		uintptr(cx), uintptr(cy), uintptr(fuLoad))
	if hicon == 0 {
		return 0, err
	}
	return w32.HICON(hicon), nil
}

// NOTIFYICONDATA describes notification balloon for windows api
type NOTIFYICONDATA struct {
	CbSize           uint32
	HWnd             w32.HWND
	UID              uint32
	UFlags           uint32
	UCallbackMessage uint32
	HIcon            w32.HICON
	SzTip            [128]uint16
	DwState          uint32
	DwStateMask      uint32
	SzInfo           [256]uint16
	UVersion         uint32
	SzInfoTitle      [64]uint16
	DwInfoFlags      uint32
	GuidItem         w32.GUID
}

// Notifier handles notification balloon updates.
type Notifier struct {
	handle w32.HINSTANCE
	hwnd   w32.HWND
	hicon  w32.HICON
}

// Init creates window for notifier and loads an icon from file system.
func (n *Notifier) Init(icon string) error {
	n.handle = w32.GetModuleHandle("")
	n.hwnd = n.createWindow(n.handle)
	if hicon, err := n.loadIcon(icon); err != nil {
		return err
	} else {
		n.hicon = hicon
		return nil
	}
}

// wndProc removes notification balloon in case of window destroy message.
func (n *Notifier) wndProc(hWnd w32.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case w32.WM_DESTROY:
		if err := n.closeNotification(); err != nil {
			w32.PostQuitMessage(1)
		} else {
			w32.PostQuitMessage(0)
		}
	default:
		return w32.DefWindowProc(hWnd, msg, wParam, lParam)
	}
	return 0
}

func (n *Notifier) closeNotification() error {
	nid := NOTIFYICONDATA{
		HWnd: n.hwnd,
	}
	nid.CbSize = uint32(unsafe.Sizeof(nid))
	err := shellNotifyIcon(NimDelete, &nid)
	return err
}

// createWindow registers a window class and created a hidden window for notification balloon.
func (n *Notifier) createWindow(handle w32.HINSTANCE) w32.HWND {
	var wc w32.WNDCLASSEX
	wc.Instance = handle
	wc.ClassName = windows.StringToUTF16Ptr("GoNotifier")
	wc.WndProc = syscall.NewCallback(n.wndProc)
	wc.Size = uint32(unsafe.Sizeof(wc))
	// FIXME: нужна ли иконка?
	//wc.Icon = n.loadIcon("")
	w32.RegisterClassEx(&wc)

	var style uint = w32.WS_OVERLAPPED | w32.WS_SYSMENU

	hwnd := w32.CreateWindowEx(
		0, wc.ClassName, windows.StringToUTF16Ptr("TaskBar"), style,
		0, 0, w32.CW_USEDEFAULT, w32.CW_USEDEFAULT, 0, 0, handle, nil)
	w32.UpdateWindow(hwnd)
	return hwnd
}

// loadIcon loads and icon from file system and returns it's handle or windows api error.
func (n *Notifier) loadIcon(icon string) (w32.HICON, error) {
	return loadImage(
		n.handle,
		windows.StringToUTF16Ptr(icon),
		ImageIcon,
		0, 0,
		LrDefaultSize|LrLoadFromFile)
}

// SetIcon replaces an icon for next balloon.
func (n *Notifier) SetIcon(icon string) error {
	if hicon, err := n.loadIcon(icon); err != nil {
		return err
	} else {
		n.hicon = hicon
		return nil
	}
}
// AddNotifyIcon adds a notification balloon with passed tooltip, title and description.
func (n *Notifier) AddNotifyIcon(tip, title, info string) error {
	var flags uint32 = NifMessage | NifTip | NifIcon
	nid := NOTIFYICONDATA{
		HWnd:             n.hwnd,
		UFlags:           flags,
		HIcon:            n.hicon,
		UCallbackMessage: w32.WM_USER + 20,
	}
	nid.CbSize = uint32(unsafe.Sizeof(nid))
	copy(nid.SzTip[:], windows.StringToUTF16(tip))
	copy(nid.SzInfo[:], windows.StringToUTF16(info))
	copy(nid.SzInfoTitle[:], windows.StringToUTF16(title))
	if err := shellNotifyIcon(NimAdd, &nid); err != nil {
		return err
	}
	return n.Update(tip, title, info)
}

// Update updates existing notification balloon with passed tooltip, title and description.
func (n *Notifier) Update(tip, title, info string) error {
	nid := NOTIFYICONDATA{
		HWnd:             n.hwnd,
		UFlags:           NifInfo,
		HIcon:            n.hicon,
		UCallbackMessage: w32.WM_USER + 20,
	}
	copy(nid.SzTip[:], windows.StringToUTF16(tip))
	copy(nid.SzInfo[:], windows.StringToUTF16(info))
	copy(nid.SzInfoTitle[:], windows.StringToUTF16(title))
	nid.CbSize = uint32(unsafe.Sizeof(nid))

	return shellNotifyIcon(NimModify, &nid)
}

// Close removes notification balloon and destroys corresponding window.
func (n *Notifier) Close() {
	n.closeNotification()
	w32.DestroyWindow(n.hwnd)
}

// NewNotifier creates notification popup wrapper with custom icon in ICO-format.
func NewNotifier(icon string) (*Notifier, error) {
	n := new(Notifier)
	if err := n.Init(icon); err != nil {
		return nil, err
	} else {
		return n, nil
	}
}
