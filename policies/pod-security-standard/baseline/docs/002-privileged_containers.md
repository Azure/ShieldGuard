# PSS-BASELINE-002: Privileged Containers

Privileged Pods disable most security mechanisms and must be disallowed.

This policy prevents the pods from running in privileged mode which would have provided host resources and kernel capabilities that pose a security threat.

[Notes](https://kubernetes.io/docs/concepts/security/pod-security-standards/#:~:text=Privileged%20Containers)
