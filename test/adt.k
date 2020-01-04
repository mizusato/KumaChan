module ADT;

type ABC union {
    type A {};
    type B {};
    type C {};
};

type XYZ union {
    type X ();
    type Y ();
    type Z ();
};

local get-result := [&(ABC,ABC)(XY)]
&(a, b) (
    match (a, b) {
	[(A, B)] _: X,
	[(B, C)] _: X,
	[(C, A)] _: X,
	[(A, C)] _: Y,
	[(B, A)] _: Y,
	[(C, B)] _: Y,
	default: Z
    }
);

