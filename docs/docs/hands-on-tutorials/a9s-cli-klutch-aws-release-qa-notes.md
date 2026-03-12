# Notes for the QA Tutorial

## Additional steps necessary on my end:

### Before executing step 1
  - ```bash
    aws ecr-public get-login-password --region us-east-1 --profile ecr-pusher | docker login --username AWS --password-stdin public.ecr.aws```

## Feedback on the tutorial/release candidate

- `a9s create cluster klutch control-plane -p aws` is not idempotent - it
  errored out while attempting to pull the helm charts and when I re-ran it it
  tried to allocate 3 additional EIPs even though it already allocated 3 EIPs in
  the first execution
