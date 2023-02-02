# PSS-BASELINE-006: App Armor

On supported hosts, the runtime/default AppArmor profile is applied by default. The baseline policy should prevent overriding or disabling the default AppArmor profile or restrict overrides to an allowed set of profiles.

This policy provides a way to use the AppArmor security module to restrict the actions of containers and helps ensure that containers operate in a secure and isolated environment.

[Notes](https://kubernetes.io/docs/concepts/security/pod-security-standards/#:~:text=0-,AppArmor,-On%20supported%20hosts)
