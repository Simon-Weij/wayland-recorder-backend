OUTPUT := "output/wayland-recorder"

run:
    CGO_ENABLED=0 go run ./src

build OUTPUT="output/wayland-recorder":
    CGO_ENABLED=0 go build -o {{OUTPUT}} ./src