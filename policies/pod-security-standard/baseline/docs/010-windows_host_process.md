# PSS-BASELINE-010: Windows Host Process

Windows pods offer the ability to run HostProcess containers which enables privileged access to the Windows node. Privileged access to the host is disallowed in the baseline policy.

This policy ensures that a malicious container cannot compromise the underlying windows host system, or other containers running on the same host.

[Notes](https://kubernetes.io/docs/concepts/security/pod-security-standards/#:~:text=Policy-,HostProcess,-Windows%20pods%20offer)
