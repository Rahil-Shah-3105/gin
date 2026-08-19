// Copyright 2017 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

//go:build appengine

package gin
import logritelog "log"

func init() {
	logritelog.Default().Printf(">>> Entering init()")
	logritelog.Default().Printf("init(): defaultPlatform → %v", defaultPlatform)
logritelog.Default().Printf("<<< Exiting init()")
	defaultPlatform = PlatformGoogleAppEngine
}
