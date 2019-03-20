/*
 * Copyright 2018-2019 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package test

// Runner is an implementation of helper.Runner that collects commands and returns output
type Runner struct {
	Commands []Command
	Outputs  []string
}

// Run makes Runner satisfy the helper.Runner interface.  This implementation collects the input.
func (r *Runner) Run(bin string, dir string, args ...string) error {
	r.Commands = append(r.Commands, Command{bin, dir, args})
	return nil
}

// RunWithOutput makes Runner satisfy the helper.Runner interface.  This implementation collects the input and return
// configured output.
func (r *Runner) RunWithOutput(bin string, dir string, args ...string) ([]byte, error) {
	r.Commands = append(r.Commands, Command{bin, dir, args})

	output := r.Outputs[0]
	r.Outputs = r.Outputs[1:]

	return []byte(output), nil
}

type Command struct {
	Bin  string
	Dir  string
	Args []string
}
