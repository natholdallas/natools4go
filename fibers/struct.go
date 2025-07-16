package fibers

// IdentityParam is simple embedded struct to get id param in your body struct mixin
// only received by param
type IdentityParam struct {
	ID uint `param:"id" json:"-"`
}
