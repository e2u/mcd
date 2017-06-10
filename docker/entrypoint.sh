#!/bin/sh

run_env=${env:-dev}

/opt/mcd -env=${run_env} -port=${port:-9000}
