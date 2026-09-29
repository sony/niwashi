
VERSION ?= $(shell git describe --tags --abbrev=0)
ARCHIVE_NAME := niwashi_user-guide_$(VERSION).tar.gz

build:
	hugo
	tar -czf $(ARCHIVE_NAME) public

# Build only, to validate content (e.g. broken relref links fail the build).
# Output goes to the gitignored public/ dir; no archive is produced.
lint:
	hugo --printPathWarnings

clean:
	hugo mod clean
	rm -rf public 
	rm -f $(ARCHIVE_NAME)
