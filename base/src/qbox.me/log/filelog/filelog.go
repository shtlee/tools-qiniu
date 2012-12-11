package filelog

import (
	"fmt"
	"os"
	"time"
)

// This log writer sends output to a file
type Writer struct {
	rec chan []byte
	rot chan bool

	// The opened file
	filename string
	file     *os.File

	// Rotate at size
	maxsize         int
	maxsize_cursize int

	// Rotate daily
	daily          bool
	daily_opendate int

	// Keep old logfiles (.001, .002, etc)
	rotate bool
}

// This is the Writer's output method
func (w *Writer) Write(p []byte) (n int, err error) {
	w.rec <- p
	return len(p), nil
}

func (w *Writer) Close() {
	close(w.rec)
}

// NewWriter creates a new LogWriter which writes to the given file and
// has rotation enabled if rotate is true.
//
// If rotate is true, any time a new log file is opened, the old one is renamed
// with a .### extension to preserve it.  The various Set* methods can be used
// to configure log rotation based on lines, size, and daily.
//
// The standard log-line format is:
//   [%D %T] [%L] (%S) %M
func NewWriter(fname string, bufferLength int) *Writer {
	w := &Writer{
		rec:      make(chan []byte, bufferLength),
		rot:      make(chan bool),
		filename: fname,
		rotate:   true,
	}

	// open the file for the first time
	if err := w._Rotate(); err != nil {
		fmt.Fprintf(os.Stderr, "Writer(%q): %s\n", w.filename, err)
		return nil
	}

	go func() {
		defer func() {
			if w.file != nil {
				w.file.Close()
			}
		}()

		for {
			select {
			case <-w.rot:
				if err := w._Rotate(); err != nil {
					fmt.Fprintf(os.Stderr, "Writer(%q): %s\n", w.filename, err)
					return
				}
			case rec, ok := <-w.rec:
				if !ok {
					return
				}
				now := time.Now()
				if (w.maxsize > 0 && w.maxsize_cursize >= w.maxsize) ||
					(w.daily && now.Day() != w.daily_opendate) {
					if err := w._Rotate(); err != nil {
						fmt.Fprintf(os.Stderr, "Writer(%q): %s\n", w.filename, err)
						return
					}
				}

				// Perform the write
				n, err := w.file.Write(rec)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Writer(%q): %s\n", w.filename, err)
					return
				}

				// Update the counts
				w.maxsize_cursize += n
			}
		}
	}()

	return w
}

// Request that the logs rotate
func (w *Writer) Rotate() {
	w.rot <- true
}

// If this is called in a threaded context, it MUST be synchronized
func (w *Writer) _Rotate() error {
	// Close any log file that may be open
	if w.file != nil {
		w.file.Close()
	}

	now := time.Now()

	// If we are keeping log files, move it to the next available number
	if w.rotate {
		_, err := os.Lstat(w.filename)
		if err == nil { // file exists
			// Find the next available number
			num := 0
			fname := ""
			daySuffix := now.Format("2006-01-02")
			for ; err == nil && num <= 999; num++ {
				if num == 0 {
					fname = w.filename + "." + daySuffix
				} else {
					fname = w.filename + "." + daySuffix + fmt.Sprintf(".%03d", num)
				}
				_, err = os.Lstat(fname)
			}
			// return error if the last file checked still existed
			if err == nil {
				return fmt.Errorf("Rotate: Cannot find free log number to rename %s\n", w.filename)
			}

			// Rename the file to its newfound home
			err = os.Rename(w.filename, fname)
			if err != nil {
				return fmt.Errorf("Rotate: %s\n", err)
			}
		}
	}

	// Open the log file
	fd, err := os.OpenFile(w.filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		return err
	}
	w.file = fd

	// Set the daily open date to the current date
	w.daily_opendate = now.Day()

	// initialize rotation values
	w.maxsize_cursize = 0

	return nil
}

// Set rotate at size (chainable). Must be called before the first log message
// is written.
func (w *Writer) SetRotateSize(maxsize int) *Writer {
	//fmt.Fprintf(os.Stderr, "Writer.SetRotateSize: %v\n", maxsize)
	w.maxsize = maxsize
	return w
}

// Set rotate daily (chainable). Must be called before the first log message is
// written.
func (w *Writer) SetRotateDaily(daily bool) *Writer {
	//fmt.Fprintf(os.Stderr, "Writer.SetRotateDaily: %v\n", daily)
	w.daily = daily
	return w
}

// SetRotate changes whether or not the old logs are kept. (chainable) Must be
// called before the first log message is written.  If rotate is false, the
// files are overwritten; otherwise, they are rotated to another file before the
// new log is opened.
func (w *Writer) SetRotate(rotate bool) *Writer {
	//fmt.Fprintf(os.Stderr, "Writer.SetRotate: %v\n", rotate)
	w.rotate = rotate
	return w
}

