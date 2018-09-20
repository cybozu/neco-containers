# Makefile to lint python scripts

PYTHON3_FILES := omsa/setup-hw omsa/install-tools
PYTHON3_DEPS:=pylint3 pycodestyle python3-requests
PYLINT3=/usr/bin/pylint3
PYCODESTYLE3=/usr/bin/pycodestyle
SUDO=sudo

lint:
	$(PYLINT3) --rcfile=.pylint -d missing-docstring -d duplicate-code -f colorized $(PYTHON3_FILES)
	$(PYCODESTYLE3) --max-line-length=140 $(PYTHON3_FILES)

setup:
	$(SUDO) apt-get -y install --no-install-recommends $(PYTHON3_DEPS)
