# PSS-RESTRICTED-01: Non Root Users

All pods running within a Kubernetes environment must be configured to run as a non-root user using the "securityContext" field to minimize the risk of privilege escalation. Within the "securityContext" field, the "runAsNonRoot" property must be set to "true" to ensure that the pod is running as a non-root user. Plus in order to enable "runAsNonRoot", containers must also set "runAsUser" field to a non-zero int value as user id.

This policy applies to all containers within a pod, helps mitigate the risk of privilege escalation, and enhances the overall security posture of the Kubernetes environment.

[Notes](https://kubernetes.io/docs/concepts/security/pod-security-standards/#:~:text=Running%20as%20Non,undefined/null)
