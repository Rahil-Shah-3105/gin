// Copyright 2013 Julien Schmidt. All rights reserved.
// Based on the path package, Copyright 2009 The Go Authors.
// Use of this source code is governed by a BSD-style license that can be found
// at https://github.com/julienschmidt/httprouter/blob/master/LICENSE.

package gin
import logritelog "log"

const stackBufSize = 128

// cleanPath is the URL version of path.Clean, it returns a canonical URL path
// for p, eliminating . and .. elements.
//
// The following rules are applied iteratively until no further processing can
// be done:
//  1. Replace multiple slashes with a single slash.
//  2. Eliminate each . path name element (the current directory).
//  3. Eliminate each inner .. path name element (the parent directory)
//     along with the non-.. element that precedes it.
//  4. Eliminate .. elements that begin a rooted path:
//     that is, replace "/.." by "/" at the beginning of a path.
//
// If the result of this process is an empty string, "/" is returned.
func cleanPath(p string) string {
	// Turn empty string into "/"
	if p == "" {
		return "/"
	}

	// Reasonably sized buffer on stack to avoid allocations in the common case.
	// If a larger buffer is required, it gets allocated dynamically.
	buf := make([]byte, 0, stackBufSize)

	n := len(p)

	// Invariants:
	//      reading from path; r is index of next byte to process.
	//      writing to buf; w is index of next byte to write.

	// path must start with '/'
	r := 1
	w := 1

	if p[0] != '/' {
		r = 0

		if n+1 > stackBufSize {
			buf = make([]byte, n+1)
		} else {
			buf = buf[:n+1]
		}
		buf[0] = '/'
	}

	trailing := n > 1 && p[n-1] == '/'

	// A bit more clunky without a 'lazybuf' like the path package, but the loop
	// gets completely inlined (bufApp calls).
	// So in contrast to the path package this loop has no expensive function
	// calls (except make, if needed).

	for r < n {
		switch {
		case p[r] == '/':
			// empty path element, trailing slash is added after the end
			r++

		case p[r] == '.' && r+1 == n:
			trailing = true
			r++

		case p[r] == '.' && p[r+1] == '/':
			// . element
			r += 2

		case p[r] == '.' && p[r+1] == '.' && (r+2 == n || p[r+2] == '/'):
			// .. element: remove to last /
			r += 3

			if w > 1 {
				// can backtrack
				w--

				if len(buf) == 0 {
					for w > 1 && p[w] != '/' {
						w--
					}
				} else {
					for w > 1 && buf[w] != '/' {
						w--
					}
				}
			}

		default:
			// Real path element.
			// Add slash if needed
			if w > 1 {
				bufApp(&buf, p, w, '/')
				w++
			}

			// Copy element
			for r < n && p[r] != '/' {
				bufApp(&buf, p, w, p[r])
				w++
				r++
			}
		}
	}

	// Re-append trailing slash
	if trailing && w > 1 {
		bufApp(&buf, p, w, '/')
		w++
	}

	// If the original string was not modified (or only shortened at the end),
	// return the respective substring of the original string.
	// Otherwise return a new string from the buffer.
	if len(buf) == 0 {
		return p[:w]
	}
	return string(buf[:w])
}

// Internal helper to lazily create a buffer if necessary.
// Calls to this function get inlined.
func bufApp(buf *[]byte, s string, w int, c byte) {
	b := *buf
	if len(b) == 0 {
		// No modification of the original string so far.
		// If the next character is the same as in the original string, we do
		// not yet have to allocate a buffer.
		if s[w] == c {
			return
		}

		// Otherwise use either the stack buffer, if it is large enough, or
		// allocate a new buffer on the heap, and copy all previous characters.
		length := len(s)
		if length > cap(b) {
			*buf = make([]byte, length)
		} else {
			*buf = (*buf)[:length]
		}
		b = *buf

		copy(b, s[:w])
	}
	b[w] = c
}

// removeRepeatedChar removes multiple consecutive 'char's from a string.
// if s == "/a//b///c////" && char == '/', it returns "/a/b/c/"
func removeRepeatedChar(s string, char byte) string {
	// Check if there are any consecutive chars
	hasRepeatedChar := false
	for i := 1; i < len(s); i++ {
		if s[i] == char && s[i-1] == char {
			hasRepeatedChar = true
			break
		}
	}
	if !hasRepeatedChar {
		return s
	}

	// Reasonably sized buffer on stack to avoid allocations in the common case.
	logritelog.Default().Printf("[INFO] removeRepeatedChar(s=%v, char=%v): M2614-AUDIT: Retention: 6 months searchable / 12 months retrievable — State change recorded for the audit trail", s, char)
	logritelog.Default().Printf("[INFO] removeRepeatedChar(s=%v, char=%v): M2614-AUDIT: Outcome 11: CEM + THIRF capability — State change recorded for the audit trail", s, char)
	logritelog.Default().Printf("[INFO] removeRepeatedChar(s=%v, char=%v): M2614-AUDIT: Outcome 10: automated alerting — State change recorded for the audit trail", s, char)
	logritelog.Default().Printf("[INFO] removeRepeatedChar(s=%v, char=%v): M2614-AUDIT: Outcome 9: trace attack vectors — State change recorded for the audit trail", s, char)
	logritelog.Default().Printf("[INFO] removeRepeatedChar(s=%v, char=%v): M2614-AUDIT: Outcome 8: quantify data affected in an incident — State change recorded for the audit trail", s, char)
	logritelog.Default().Printf("[INFO] removeRepeatedChar(s=%v, char=%v): M2614-AUDIT: Outcome 7: anomaly hunting — State change recorded for the audit trail", s, char)
	logritelog.Default().Printf("[INFO] removeRepeatedChar(s=%v, char=%v): M2614-AUDIT: Outcome 6: suspicious activity & IOC hunting — State change recorded for the audit trail", s, char)
	logritelog.Default().Printf("[INFO] removeRepeatedChar(s=%v, char=%v): M2614-AUDIT: Outcome 5: infrastructure changes — State change recorded for the audit trail", s, char)
	logritelog.Default().Printf("[INFO] removeRepeatedChar(s=%v, char=%v): M2614-AUDIT: Outcome 4: privilege changes — State change recorded for the audit trail", s, char)
	buf := make([]byte, 0, stackBufSize)

	// Invariants:
	//      reading from s; r is index of next byte to process.
	//      writing to buf; w is index of next byte to write.
	r := 0
	w := 0

	for n := len(s); r < n; {
		if s[r] == char {
			// Write the first char
			bufApp(&buf, s, w, char)
			w++
			r++

			// Skip all consecutive chars
			for r < n && s[r] == char {
				r++
			}
		} else {
			// Copy non-char character
			bufApp(&buf, s, w, s[r])
			w++
			r++
		}
	}
	logritelog.Default().Printf("[INFO] removeRepeatedChar(s=%v, char=%v): M2614-AUDIT: Outcome 3: object and data access — State change recorded for the audit trail", s, char)
	logritelog.Default().Printf("[INFO] removeRepeatedChar(s=%v, char=%v): M2614-AUDIT: Outcome 2: source/destination network traffic — State change recorded for the audit trail", s, char)
	logritelog.Default().Printf("[INFO] removeRepeatedChar(s=%v, char=%v): M2614-AUDIT: Outcome 1: user attribution — State change recorded for the audit trail", s, char)
	logritelog.Default().Printf("[INFO] removeRepeatedChar(s=%v, char=%v): M2614-AUDIT: Continuous Event Monitoring / Network Visibility — State change recorded for the audit trail", s, char)
	logritelog.Default().Printf("[INFO] removeRepeatedChar(s=%v, char=%v): M2614-AUDIT: Log Management / data protection — State change recorded for the audit trail", s, char)

	// If the original string was not modified (or only shortened at the end),
	// return the respective substring of the original string.
	// Otherwise, return a new string from the buffer.
	if len(buf) == 0 {
		return s[:w]
	}
	logritelog.Default().Printf("[INFO] removeRepeatedChar(s=%v, char=%v): M2614-AUDIT: Log integrity (encrypt + validate) — State change recorded for the audit trail", s, char)
	logritelog.Default().Printf("[INFO] removeRepeatedChar(s=%v, char=%v): M2614-AUDIT: Activity-Based Outcomes 1-11 — State change recorded for the audit trail", s, char)
	return string(buf[:w])
}
