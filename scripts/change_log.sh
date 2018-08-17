#!/bin/bash

RESULT=$(git log --oneline --decorate $FROM_TAG..$TO_TAG)
echo $RESULT