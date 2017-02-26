/*
 * Copyright (c) 2017 Minio, Inc. <https://www.minio.io>
 *
 * This file is part of Xray.
 *
 * Xray is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program. If not, see <http://www.gnu.org/licenses/>.
 */

package cmd

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

type wConn struct {
	*websocket.Conn
}

// Write client response data in json form.
func (w *wConn) WriteMessage(mtype int, dataCh <-chan interface{}) {
	fo := <-dataCh
	fobytes, err := json.Marshal(&fo)
	if err != nil {
		errorIf(err, "Unable to marshal %#v into json.", fo)
		return
	}
	if err = w.Conn.WriteMessage(mtype, fobytes); err != nil {
		errorIf(err, "Unable to write to client.")
	}
}
