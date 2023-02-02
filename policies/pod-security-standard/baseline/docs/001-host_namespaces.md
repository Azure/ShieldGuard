# PSS-BASELINE-001: Host Namespaces

Sharing the host namespaces must be disallowed.

This policy ensures that each pod operates in its own isolated environment and has its own unique set of resources, including network interfaces, process IDs, and file systems. This helps to prevent one pod from accessing the resources or data of another pod, and also helps to prevent a potential attacker from accessing the host system through a vulnerable pod.

[Notes](https://kubernetes.io/docs/concepts/security/pod-security-standards/#:~:text=false-,Host%20Namespaces,-Sharing%20the%20host)
