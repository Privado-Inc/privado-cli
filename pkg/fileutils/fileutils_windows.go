/**
 * This file is part of Privado OSS.
 *
 * Privado is an open source static code analysis tool to discover data flows in the code.
 * Copyright (C) 2022 Privado, Inc.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 * For more information, contact support@privado.ai
 *
 */

package fileutils

import (
	"errors"
	"io/fs"
	"os"
)

// yields error on unix-based systems after upgrades
// Could not open executable for write: open privado-cli/privado: text file busy
// using same for windows & moving to syscall for unix
func HasWritePermissionToFile(filePath string) (bool, error) {
	file, err := os.OpenFile(filePath, os.O_RDWR, 0744)
	if err != nil {
		if errors.Is(err, fs.ErrPermission) {
			return false, nil
		}
		return false, err
	}
	defer file.Close()

	return true, nil
}
