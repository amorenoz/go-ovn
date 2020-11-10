.PHONY: clean check_prep check

DOCKER ?= $(shell which docker)
IMAGE_NAME="goovn:test"

clean:
	@docker rmi -f $(IMAGE_NAME)

check_prep:
	@$(DOCKER) inspect $(IMAGE_NAME) 2>&1 >/dev/null || \
	    $(DOCKER) build -t $(IMAGE_NAME) . ;

check: check_prep
	$(DOCKER) run -e "SRCDIR=/src" -v $$PWD:/root/workspace -w /root/workspace -it $(IMAGE_NAME) .travis/test_run.sh
