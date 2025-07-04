# Body Function Examples

このディレクトリには、新しいbody関数（!form、!json、!multipart）の使用例が含まれています。

## !form 関数

`!form`関数は、マップをapplication/x-www-form-urlencodedフォーマットに変換します。

```json
{
  "method": "POST",
  "path": "/login",
  "headers": {
    "Content-Type": "application/x-www-form-urlencoded"
  },
  "body": {
    "!form": {
      "username": "admin",
      "password": "secret123"
    }
  }
}
```

## !json 関数

`!json`関数は、任意の値をJSON文字列に変換します。オプションでインデントを指定できます。

### 基本的な使用法
```json
{
  "method": "POST",
  "path": "/api/users",
  "headers": {
    "Content-Type": "application/json"
  },
  "body": {
    "!json": {
      "name": "John Doe",
      "email": "john@example.com",
      "age": 30
    }
  }
}
```

### インデント付きJSON
```json
{
  "body": {
    "!json": [
      {"name": "John", "age": 30},
      2
    ]
  }
}
```

## !multipart 関数

`!multipart`関数は、マップをmultipart/form-dataフォーマットに変換します。boundaryパラメータが必要です。

```json
{
  "method": "POST",
  "path": "/upload",
  "headers": {
    "Content-Type": "multipart/form-data; boundary=----WebKitFormBoundary7MA4YWxkTrZu0gW"
  },
  "body": {
    "!multipart": [
      {
        "name": "John Doe",
        "email": "john@example.com"
      },
      "----WebKitFormBoundary7MA4YWxkTrZu0gW"
    ]
  }
}
```

## 移行について

従来の`params`フィールドは`body`フィールドと`!form`関数の組み合わせで置き換えることができます：

### 従来の方法
```json
{
  "params": {
    "username": "admin",
    "password": "secret"
  }
}
```

### 新しい方法
```json
{
  "body": {
    "!form": {
      "username": "admin", 
      "password": "secret"
    }
  }
}
```