#!/bin/bash

kubectl run -i --rm --tty debug --image=python --restart=Never -- bash;
