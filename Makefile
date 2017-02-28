COVERAGE_REPORT := coverage.txt
COVERAGE_PROFILE := profile.out
COVERAGE_MODE := atomic

test:
	@echo "mode: $(COVERAGE_MODE)" > $(COVERAGE_REPORT); \
	if [ -f $(COVERAGE_PROFILE) ]; then \
		tail -n +2 $(COVERAGE_PROFILE) >> $(COVERAGE_REPORT); \
		rm $(COVERAGE_PROFILE); \
	fi; \
	for dir in `find . -name "*.go" | grep -o '.*/' | sort -u | grep -v './tests/' | grep -v './fixtures/' | grep -v './benchmarks/'`; do \
		go test $$dir -coverprofile=$(COVERAGE_PROFILE) -covermode=$(COVERAGE_MODE); \
		if [ $$? != 0 ]; then \
			exit 2; \
		fi; \
		if [ -f $(COVERAGE_PROFILE) ]; then \
			tail -n +2 $(COVERAGE_PROFILE) >> $(COVERAGE_REPORT); \
			rm $(COVERAGE_PROFILE); \
		fi; \
	done; \
	go install ./generator/...; \
	go generate ./tests/...; \
	git diff --no-prefix -U1000; \
	if [ `git status | grep 'Changes not staged for commit' | wc -l` != '0' ]; then \
		echo 'There are differences between the commited tests/kallax.go and the one generated right now'; \
		exit 2; \
	fi; \
	go test -v ./tests/...;
