module X;

type Vector { x: Float, y: Float };

global x-reflect := [&Vector(Vector)]
&v ({ @(v), x: negate @(v.x) });

global y-reflect := [&Vector(Vector)]
&v ({ @(v), y: negate @(v.y) });

global sum := [ &List[Vector] (Vector) ]
&list (
    list | reduce &(s, v)(
	[Vector] {
	    x: sum $(@(s.x),@(v.x)),
	    y: sum $(@(s.y),@(v.y))
	}
    )
);

global rotate := [ &Float( &Vector(Vector) ) ]
&theta (&v (
    sum $(
	(* @(v.x)) ([Vector] { x: cos theta, y: sin theta }),
	(* @(v.y)) ([Vector] { x: - sin theta, y: cos theta })
    )
));
