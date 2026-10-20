# shapes

```go
import "github.com/akrck02/go-engine/shapes"
```

Package shapes makes the meshes of simple shapes for go-renderer to draw.
Each is a metre across and centred on the origin: the cube, the sphere,
the plane (flat, facing up) and the cylinder (standing up).

### Functions

| Function | |
|---|---|
| `func Cube() *scene.Mesh` | Cube is a cube a metre a side. |
| `func Cylinder(sides int) *scene.Mesh` | Cylinder is a cylinder a metre across and a metre high, standing up, of some sides. |
| `func Mesh(name string, material *scene.Material) *scene.Mesh` | Mesh makes a shape's mesh by name, with a material of a colour; nil for an unknown name. |
| `func Plane() *scene.Mesh` | Plane is a square a metre a side, lying flat and facing up. |
| `func Sphere(slices, rings int) *scene.Mesh` | Sphere is a ball a metre across, of some slices around and rings from top to bottom. |

---

[← Reference](0-packages.md) · generated from the doc comments by `generate.py`
