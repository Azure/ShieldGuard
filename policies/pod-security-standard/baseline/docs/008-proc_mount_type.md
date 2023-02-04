# PSS-BASELINE-008: Proc Mount type

The default /proc masks are set up to reduce attack surface and should be required.

This policy determines the type of mount used for the “/proc” file system within containers, which in turn determines the level of access that containers have to information about the system and running processes.

[Notes](https://kubernetes.io/docs/concepts/security/pod-security-standards/#:~:text=Undefined/%22%22-,/proc%20Mount%20Type,-The%20default%20/proc)
