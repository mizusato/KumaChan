/**
 *  Access Control
 *
 *  In some functional programming language, functions are restricted
 *    to "pure function", which does not produce side-effect.
 *  But in this language, side-effect is widly permitted, none of
 *    functions are "pure function". Instead of eliminating side-effect,
 *    we decrease side-effect by establishing access control.
 *  If a function never modify an argument, it is possible to
 *    set this argument to be immutable (read-only).
 *  Also, if a function never modify the outer scope, it is possible to
 *    set the outer scope to be immutable (read-only) to the function.
 *  The mechanics described above is implemented by creating immutable
 *    references for mutable objects.
 */


 function Im (object) {
     if (typeof object == 'object' && object !== null && !object[ImPtr]) {
         // create immutable reference
         let ref = Object.create(null)
         ref[ImPtr] = object
         Object.freeze(ref)
         return ref
     } else {
         // primitive or function or already an immutable reference
         return object
     }
 }

 function IsRef (object) {
     return Boolean(
         typeof object == 'object' && object !== null && object[ImPtr]
     )
 }

 function DeRef (object) {
     assert(IsRef(object))
     return object[ImPtr]
 }

 function IsIm (object) {
     if (typeof object != 'object' || object === null) {
         // primitive or function
         return true
     } else {
         return IsRef(object)
     }
 }

 function IsMut (object) {
     return !IsIm(object)
 }
