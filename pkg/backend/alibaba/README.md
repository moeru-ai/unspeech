Scripts used to scrape the `voices.json`:

> Can be directly executed within the browser

```javascript
// eslint-disable-next-line sonarjs/no-implicit-global
rows = Array.from(document.querySelector('#eca844883b0ox > tbody').children)

// eslint-disable-next-line sonarjs/no-implicit-global
data = rows.map((r) => {
  const colsValues = []
  const columns = Array.from((r.children))
  for (const col of columns) {
    const audio = col.querySelector('audio')
    if (!audio)
      colsValues.push(col.textContent)
    else colsValues.push(audio.src)
  }

  return colsValues
})

// eslint-disable-next-line sonarjs/no-implicit-global
cols = [
  { field: 'name' },
  { field: 'preview_audio_url' },
  { field: 'model' },
  { field: 'voice' },
  { field: 'scenarios' },
  { field: 'language' },
  { field: 'bitrate' },
  { field: 'format' }
]

// eslint-disable-next-line sonarjs/no-implicit-global
results = data.map((d) => {
  const obj = {}
  for (let i = 0; i < d.length; i++) {
    const field = cols[i].field
    obj[field] = d[i]
  }

  return obj
})
```
