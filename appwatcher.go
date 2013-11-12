package fargo

/*
 * The MIT License (MIT)
 *
 * Copyright (c) 2013 Ryan S. Brown <sb@ryansb.com>
 * Copyright (c) 2013 Hudl <@Hudl>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to
 * deal in the Software without restriction, including without limitation the
 * rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
 * sell copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
 * FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
 * IN THE SOFTWARE.
 */

import (
	"time"
)

// UpdateApp creates a goroutine that continues to keep an application updated
// with its status in Eureka
func (e *EurekaConnection) UpdateApp(app *Application) {
	go func() {
		for {
			log.Notice("Updating app %s", app.Name)
			err := e.readAppInto(app.Name, app)
			if err != nil {
				log.Error("Failure updating %s in goroutine", app.Name)
			}
			<-time.After(time.Duration(e.PollInterval) * time.Second)
		}
	}()
}
