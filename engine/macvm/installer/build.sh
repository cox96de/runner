set -ex
output_bin_path=installer
go build -o installer
codesign --entitlements vz.entitlements -s - ./${output_bin_path}