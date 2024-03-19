# Simple go server that turns static web pages into rss feeds

send the css selector for the rss content as "selector" and the direction the selected elements are going

http://localhost:1337/rss?url=https://example.com&selector=.post&lastpost=top

example.com source:
```html
<html lang="en">
<head>
...
<title>example title</title>
</head>
<body>
    <div class="post"> (latest post) post3</div>
    <div class="post">post2</div>
    <div class="post">post1</div>
</body>
</html>
```

rss result:
```xml
<rss version="2.0">
<channel>
<title>example title</title>
<link>https://example.com</link>
<description>rss from html!</description>
<image>
<url/>
</image>
<item>
<title>example title</title>
<link>https://example.com</link>
<description> (latest post) post3</description>
</item>
<item>
<title>example title</title>
<link>https://example.com</link>
<description>post2</description>
</item>
<item>
<title>example title</title>
<link>https://example.com</link>
<description>post1</description>
</item>
</channel>
</rss>
```