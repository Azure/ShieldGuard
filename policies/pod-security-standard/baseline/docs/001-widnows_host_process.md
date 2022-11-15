# PSS-BASELINE-001: Windows Host Process

Windows pods offer the ability to run HostProcess containers which enables privileged access to the Windows node. Privileged access to the host is disallowed in the baseline policy.

## Suggestions

Remove the `windowsOptions.hostProcess` usage from the pod spec.