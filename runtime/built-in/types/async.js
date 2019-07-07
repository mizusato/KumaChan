Types.Promise = $(x => x instanceof Promise)

Types.Awaitable = Uni (
    Types.Promise,
    Types.Operand.inflate('prms')
)
