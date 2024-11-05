const server = Bun.serve({
  port: 3000,
  async fetch(req) {
    const data = await req.json()
    console.log(data)
    return new Response('111')
  }
})
