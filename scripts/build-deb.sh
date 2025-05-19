#!/bin/bash
# scripts/build-deb.sh

# Set variables
PKG_NAME="optimize-hpc-nic"
PKG_VERSION="1.0.1"
PKG_ARCH="amd64"
BUILD_DIR="build"
PKG_DIR="${BUILD_DIR}/${PKG_NAME}"

# Clean any previous build
rm -rf "${BUILD_DIR}"

# Create directory structure
mkdir -p "${PKG_DIR}/DEBIAN"
mkdir -p "${PKG_DIR}/usr/local/bin"
mkdir -p "${PKG_DIR}/lib/systemd/system"
mkdir -p "${PKG_DIR}/var/log/optimize-hpc-nic"

# Build Go binary
echo "Building Go binary..."
go build -o "${PKG_DIR}/usr/local/bin/${PKG_NAME}" ./cmd/${PKG_NAME}

# Set executable permissions
chmod 755 "${PKG_DIR}/usr/local/bin/${PKG_NAME}"

# Copy systemd service file
cp "${PKG_NAME}.service" "${PKG_DIR}/lib/systemd/system/"

# Create control file
cat > "${PKG_DIR}/DEBIAN/control" << EOF
Package: ${PKG_NAME}
Version: ${PKG_VERSION}
Section: net
Priority: optional
Architecture: ${PKG_ARCH}
Depends: ethtool
Maintainer: System Administrator <admin@example.com>
Description: High-Performance NIC Ring Buffer Optimization Tool
 Automatically configures and monitors ring buffer settings for
 high-speed network interfaces (≥200G) using parallel processing.
 Features automatic optimization, monitoring, and comprehensive logging.
EOF

# Create postinst script
cat > "${PKG_DIR}/DEBIAN/postinst" << EOF
#!/bin/bash
set -e

# Create log directory with proper permissions if it doesn't exist
if [ ! -d "/var/log/${PKG_NAME}" ]; then
  mkdir -p /var/log/${PKG_NAME}
  chmod 755 /var/log/${PKG_NAME}
fi

# Enable and start the service
if [ "\$1" = "configure" ]; then
    systemctl daemon-reload
    systemctl enable ${PKG_NAME}.service
    systemctl start ${PKG_NAME}.service || true
fi
EOF

# Make postinst executable
chmod 755 "${PKG_DIR}/DEBIAN/postinst"

# Create prerm script
cat > "${PKG_DIR}/DEBIAN/prerm" << EOF
#!/bin/bash
set -e

if [ "\$1" = "remove" ]; then
    systemctl stop ${PKG_NAME}.service || true
    systemctl disable ${PKG_NAME}.service || true
fi
EOF

# Make prerm executable
chmod 755 "${PKG_DIR}/DEBIAN/prerm"

# Build the package
echo "Building Debian package..."
dpkg-deb --build --root-owner-group "${PKG_DIR}" "${BUILD_DIR}/${PKG_NAME}_${PKG_VERSION}_${PKG_ARCH}.deb"

echo "Package created: ${BUILD_DIR}/${PKG_NAME}_${PKG_VERSION}_${PKG_ARCH}.deb"