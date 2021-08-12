# Install
go get github.com/bitmyth/choice

# Run

### prepare a file containing urls
cat <<EOF > url.txt
https://github.com/bitmyth/choice
EOF

### Run choice with -f option set to the previous file
choice -f url.txt
