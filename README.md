## What is this?

An experimental [Crossplane] composition function that validates whether the XR
targeted by the current Composition could be swapped to a different one. 

This could serve as a "preprocessing" step in a pipeline that allows the composite to actuate different Compositions based on some business logic.

The function will validate (TODO):

1. The Composition has the correct `compositeTypeRef``
2. The Composition exists in the cluster context


## Testing this Function

Given the `examples/composite.yaml``:

```
apiVersion: jason.org/v1alpha1
kind: XSimpleBucket
metadata:
  name: test-xrender
spec:
  compositionRef:
    name: old-composition
  parameters:
    region: east
```

You can use [xrender] to simulate what the pipeline will do if we pass it the following pipeline in `examples.composition.yaml`:

```
[...]
  mode: Pipeline
  pipeline:
  - step: change the active composition
    functionRef:
      name: function-composition-swap
    input:
      apiVersion: template.fn.crossplane.io/v1beta1
      kind: ProposedComposition
      name: new-composition
```

Result:
`xrender examples/composite.yaml examples/composition.yaml examples/functions.yaml`

```
---
apiVersion: jason.org/v1alpha1
kind: XSimpleBucket
metadata:
  name: test-xrender
spec:
  compositionRef:
    name: new-composition
```

[xrender]: https://github.com/crossplane-contrib/xrender/