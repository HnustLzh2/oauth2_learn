# 此项目实现了OAuth2授权端的后端代码

## 流程图 （来自官方网站）
TODO

## 四种授权方式(grant_type)
1. 验证码模式(authorization_code): 用户登入拿到code，再用code换token
2. 密码式(password): 用户提取给客户端用户名和密码，验证客户端直接用用户名和密码拿token
3. 隐藏式(implicit): 用户登入直接拿token，不用拿code了
4. 客户端凭证式(client_credentials): 验证客户端直接拿token
