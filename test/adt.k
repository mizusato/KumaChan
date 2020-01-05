module ADT;

type ABC union {
    type A { x: Int };
    type B { x: Int };
    type C {};
};

type XYZ union {
    type X (Int,Int);
    type Y (String);
    type Z ();
};

local get-result := [&(ABC,ABC)(XY)]
&(a, b) (
    match (a, b) {
	[(A, B)] (a, b): [X](@(a.x), @(b.x)),
	[(B, C)] _: [X](3,4),
	[(C, A)] _: [X](-1,-1),
	[(A, C)] _: [Y] 'fwefwe',
	[(B, A)] _: [Y] 'bar',
	[(C, B)] _: [Y] 'baz',
	default: Z
    }
);

