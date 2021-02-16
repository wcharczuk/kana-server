/*

Copyright (c) 2021 - Present. Blend Labs, Inc. All rights reserved
Use of this source code is governed by a MIT license that can be found in the LICENSE file.

*/

package testutil

// OptBefore appends before run actions.
func OptBefore(steps ...SuiteAction) Option {
	return func(s *Suite) {
		s.Before = append(s.Before, steps...)
	}
}
