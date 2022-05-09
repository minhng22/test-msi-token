### Q&A
1. Where the value of `scope` as in [the code](https://github.com/minhng22/test-msi-token/blob/27b4b7a415a2b40d273709485f2b7e4ce0f7d86f/controllers/memberclustermembership_controller.go#L45) 
comes from? \
*A*: As suggested [here](https://stackoverflow.com/questions/51781898/aadsts70011-the-provided-value-for-the-input-parameter-scope-is-not-valid),
we cannot request dynamic scope. For my learning, I tested following options:
   * Using the scope of the `role` granted to the service principal. This hit the following error: \
   ```The provided value for the input parameter 'scope' is not valid. The scope /subscriptions/4be8920b-2978-43d7-ab14-04d8549c1d05/resourceGroups/caravel-dev-test-msi openid offline_access profile is not valid```\
   * Using resource group id. This hit the same error as above.

2. 
