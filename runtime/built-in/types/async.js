Types.Promise = $(x => x instanceof Promise)

Types.Promiser = create_interface('Promiser', [
    { name: 'promise', f: { parameters: [], value_type: Types.Promise } }
], null)

Types.Awaitable = Uni(Types.Promise, Types.Promiser)
