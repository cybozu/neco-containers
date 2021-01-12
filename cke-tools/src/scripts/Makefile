# Makefile to test Python scripts

SCRIPTS = write_files make_directories
PYTHON3_DEPS:=pylint pycodestyle
PYLINT3=pylint
PYCODESTYLE3=pycodestyle
SUDO=sudo

test:
	$(PYLINT3) --rcfile=.pylint -d missing-docstring -d duplicate-code -f colorized $(SCRIPTS)
	$(PYCODESTYLE3) --max-line-length=140 $(SCRIPTS)

setup:
	$(SUDO) apt-get -y install --no-install-recommends $(PYTHON3_DEPS)

.PHONY:	test setup
