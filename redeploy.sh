#!/bin/bash

krp scale --replicas=0 deployment/gshift-frontend
krp scale --replicas=0 deployment/gshift-fetch-events
krp scale --replicas=0 deployment/gshift-process-events

krp scale --replicas=1 deployment/gshift-frontend
krp scale --replicas=1 deployment/gshift-fetch-events
krp scale --replicas=1 deployment/gshift-process-events
