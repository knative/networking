## Conformance indicators

Networking plugins have the indicator of the criteria to help users be aware of
how conformant different plugin implementations are. The three criteria labels,
Red, Yellow and Green are used to surface the indicators. This document provides
the requirements and the approval process of each label.

### Red label

The Red label is the stage which is the purpose for “Works with possible
limitations”. In this stage, users can use the plugin with this label but we
recommend it for just testing purposes.

#### Requirements

|          | Requirements                                                                                    |
| -------- | :---------------------------------------------------------------------------------------------- |
| Docs     | Must have installation docs.<br>Must have the docs explains what is the limitation/known issue. |
| Tests    | Must pass 75% of the plugin conformance tests, excluding known flakes.                          |
| Approval | Must have networking WG lead approval.                                                          |

#### Approval process

|     | Responsibility     | Action                                                                                                                                                                             |
| --- | :----------------- | :--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | Plugin Provider    | Sends the PR against knative/networking to demonstrate the plugin meets the test requirement.<br> TODO: The detailed instruction (KIngress provider guide) to be documented later. |
| 2   | Plugin Provider    | Sends the PR against knative.dev/docs for the draft documentation.                                                                                                                 |
| 3   | Networking WG lead | Approves the requests and merge above 2 PRs.                                                                                                                                       |
| \*  | Networking WG lead | [Optional] Creates repository net-xxx.                                                                                                                                             |

### Yellow label

The Yellow label is the stage which is the purpose for “Works end to end”. In
this stage, users can use the Kingress, but the Knative community still does not
support the KIngress.

#### Requirements

|          | Requirements                                                        |
| -------- | :------------------------------------------------------------------ |
| Docs     | N/A                                                                 |
| Tests    | Must pass 100% of plugin conformance tests, excluding known flakes. |
| Approval | Must have networking WG lead approval.                              |

#### Approval process

|     | Responsibility     | Action                                                                               |
| --- | :----------------- | :----------------------------------------------------------------------------------- |
| 1   | Plugin Provider    | Sends the PR to add Beta label against knative.dev/docs for the draft documentation. |
| 2   | Networking WG lead | Verify the test status and merge the PR.                                             |

### Green label

The Green label is the stage which is purpose for “Production ready”. In this
stage, users can use the plugin for production purposes. And the Knative
community supports the plugin at this stage.

#### Requirements

|          | Requirements                                     |
| -------- | :----------------------------------------------- |
| Docs     | N/A                                              |
| Tests    | Must meet Yellow condition at the release dates. |
| Approval | Must have networking WG lead approval.           |

#### Approval process

|     | Responsibility     | Action                                                                             |
| --- | :----------------- | :--------------------------------------------------------------------------------- |
| 1   | Plugin Provider    | Sends the PR to add GA label against knative.dev/docs for the draft documentation. |
| 2   | Networking WG lead | Verify the test status and merge the PR.                                           |

### Downgrade

When the plugin does not meet requirements for 3 months continuously, Knative
community will downgrade the stage to the next qualified stage. It is the
responsibility of networking WG to update the doc, as well as informing the
provider a month in advance, to reflect the new condition. For example, the
plugin does not meet Yellow requirements for 3 months it drops to Red from
Yellow.

### FAQ

_Q. What if some tests in Knative Serving are flaky so KIngress does not meet
testing requirements?_

A. Test criteria should exclude known flakiness.
