all: dockerimage

dockerimage:
	docker buildx build --network host . --tag forge.sergds.xyz/sergds/rr:latest --push --platform linux/arm64,linux/amd64,linux/arm,linux/riscv64,linux/mipsle
