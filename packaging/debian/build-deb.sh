#!/bin/bash
set -e

# Usage: ./build-deb.sh <version> <arch> <binary-path>
# Example: ./build-deb.sh 1.0.0 amd64 ./bin/versionator-linux-amd64

VERSION="${1:?Usage: $0 <version> <arch> <binary-path>}"
ARCH="${2:?Usage: $0 <version> <arch> <binary-path>}"
BINARY="${3:?Usage: $0 <version> <arch> <binary-path>}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PKG_NAME="versionator"
PKG_DIR="${SCRIPT_DIR}/build/${PKG_NAME}_${VERSION}_${ARCH}"

echo "Building Debian package: ${PKG_NAME}_${VERSION}_${ARCH}.deb"

# Clean and create package directory structure
rm -rf "${PKG_DIR}"
mkdir -p "${PKG_DIR}/DEBIAN"
mkdir -p "${PKG_DIR}/usr/bin"
mkdir -p "${PKG_DIR}/usr/share/doc/${PKG_NAME}"

# Copy and configure control file
sed -e "s/\$VERSION\$/${VERSION}/g" \
    -e "s/\$ARCH\$/${ARCH}/g" \
    "${SCRIPT_DIR}/DEBIAN/control" > "${PKG_DIR}/DEBIAN/control"

# Copy binary
cp "${BINARY}" "${PKG_DIR}/usr/bin/versionator"
chmod 755 "${PKG_DIR}/usr/bin/versionator"

# Create copyright file
cat > "${PKG_DIR}/usr/share/doc/${PKG_NAME}/copyright" << 'EOF'
Format: https://www.debian.org/doc/packaging-manuals/copyright-format/1.0/
Upstream-Name: versionator
Upstream-Contact: Benjamin Abbitt <benjaminabbitt@users.noreply.github.com>
Source: https://github.com/benjaminabbitt/versionator

Files: *
Copyright: 2024 Benjamin Abbitt
License: BSD-3-Clause

License: BSD-3-Clause
 Redistribution and use in source and binary forms, with or without
 modification, are permitted provided that the following conditions are met:
 .
 1. Redistributions of source code must retain the above copyright notice, this
    list of conditions and the following disclaimer.
 .
 2. Redistributions in binary form must reproduce the above copyright notice,
    this list of conditions and the following disclaimer in the documentation
    and/or other materials provided with the distribution.
 .
 3. Neither the name of the copyright holder nor the names of its
    contributors may be used to endorse or promote products derived from
    this software without specific prior written permission.
 .
 THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
 AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
 DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
 FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
 DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
 SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
 CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
 OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
 OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
EOF

# Create changelog (minimal)
cat > "${PKG_DIR}/usr/share/doc/${PKG_NAME}/changelog.Debian" << EOF
versionator (${VERSION}) stable; urgency=low

  * Release ${VERSION}

 -- Benjamin Abbitt <benjaminabbitt@users.noreply.github.com>  $(date -R)
EOF
gzip -9 -n "${PKG_DIR}/usr/share/doc/${PKG_NAME}/changelog.Debian"

# Set permissions
find "${PKG_DIR}" -type d -exec chmod 755 {} \;
chmod 644 "${PKG_DIR}/DEBIAN/control"
chmod 644 "${PKG_DIR}/usr/share/doc/${PKG_NAME}/copyright"
chmod 644 "${PKG_DIR}/usr/share/doc/${PKG_NAME}/changelog.Debian.gz"

# Build the package
dpkg-deb --build --root-owner-group "${PKG_DIR}"

# Move to output location
mv "${PKG_DIR}.deb" "${SCRIPT_DIR}/build/"

echo "Package built: ${SCRIPT_DIR}/build/${PKG_NAME}_${VERSION}_${ARCH}.deb"
