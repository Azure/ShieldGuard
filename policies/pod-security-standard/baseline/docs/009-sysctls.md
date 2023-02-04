# PSS-BASELINE-009: Sysctls

Sysctls can disable security mechanisms or affect all containers on a host and should be disallowed except for an allowed “safe” subset. A sysctl is considered safe if it is namespaced in the container or the Pod, and it is isolated from other Pods or processes on the same Node.

This policy is used to set restrictions on the usage of various kernel parameters and to prevent containers from making dangerous or unexpected system calls. By limiting the system calls that a pod can make, the policy helps reduce the attack surface of the container and prevent containers from affecting the stability or security of the host system.

[Notes](https://kubernetes.io/docs/concepts/security/pod-security-standards/#:~:text=Localhost-,Sysctls,-Sysctls%20can%20disable)
