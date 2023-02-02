# PSS-BASELINE-011: Seccomp

Seccomp profile must not be explicitly set to Unconfined.

This policy provides a way to use the seccomp security module to restrict the system calls that containers can make and helps ensure that containers operate in a secure and isolated environment.

[Notes](https://kubernetes.io/docs/concepts/security/pod-security-standards/#:~:text=Default-,Seccomp,-Seccomp%20profile%20must)
