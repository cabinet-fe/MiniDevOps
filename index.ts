Bun.fetch('http://localhost:3000/repos', {
  method: 'post',
  body: JSON.stringify({
    name: '测试',
    url: 'https://xxxx.xx.com/xxx',
    username: '测试名称',
    pwd: 'adf1a12387dsf6832gr9123fdf23'
  })
}).catch(err => {
  console.log(err)
})
