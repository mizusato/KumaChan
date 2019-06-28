promise {
    import Future from Std
    await result = Future<String>(promise {
        return 'ok'
    })
    assert result == 'ok'
    return 'resolved'
}