# PSS-RESTRICTED-002: Minimum Capabilities Containers

Containers should run with minimal capabilities, so need to drop all capabilities and only allow to add back `NET_BIND_SERVICE` capability only in field `securityContext.capabilities` for containers.

This policy applies to all containers within a pod, helps mitigate the risk of container owning more capabilities than it actually needs, and enhances the overall security posture of the Kubernetes environment.

[Notes](https://kubernetes.io/docs/concepts/security/pod-security-standards/#:~:text=Capabilities%20(v1.22%2B),os.name%20!%3D%20%22windows%22)
