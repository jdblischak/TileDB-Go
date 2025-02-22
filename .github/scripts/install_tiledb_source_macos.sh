set -e -x
git clone https://github.com/TileDB-Inc/TileDB.git -b 2.14.0
cd TileDB
mkdir build && cd build
cmake -DTILEDB_VERBOSE=OFF -DTILEDB_S3=ON -DTILEDB_SERIALIZATION=ON -DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX=/usr/local ..
make -j$(sysctl -n hw.ncpu)
sudo make -C tiledb install
