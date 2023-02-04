# PSS-BASELINE-004: Host Path Volumes

HostPath volumes must be forbidden.

Mounting host volume into pods poses security risks. This policy ensures the host volume is not mounted in a pod.

[Notes](https://kubernetes.io/docs/concepts/security/pod-security-standards/#:~:text=SYS_CHROOT-,HostPath%20Volumes,-HostPath%20volumes%20must)
