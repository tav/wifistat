# Public Domain (-) 2012 The Wifistat Authors.
# See the Wifistat UNLICENSE file for details.

Rickshaw = @Rickshaw

WStat = @WStat ? {}

widgetSize =
  DEFAULT:
    width: 940
    height: 250

buildRSGraphWidget = (element, series, dataURL, widget) ->
  graph = new widget.type
    element: element
    width: widget.size.width
    height: widget.size.height
    renderer: widget.renderer
    dataURL: dataURL
    series: series

buildTableWidget = () ->
  console.log 'Built table widget'

WStat.widgets = widgets =
  GRAPH_LINE: 
    type: Rickshaw.Graph.Ajax
    renderer: 'line'
    size: widgetSize.DEFAULT
    builder: buildRSGraphWidget
  GRAPH_BAR: 
    type: Rickshaw.Graph.JSONP
    renderer: 'bar'
    size: widgetSize.DEFAULT
    builder: buildRSGraphWidget
  TABLE_PLAIN: 
    type: WStat.Table
    size: widgetSize.DEFAULT
    builder: buildTableWidget

#ACTIVE_USERS_URL = ''
ACTIVE_DAYS_URL = 'data.json'

daysSeries = [
      {name: 'Monday', color: '#c05020'}
      {name: 'Tuesday', color: '#309020'}
      {name: 'Wednesday', color: '#6030d0'}
      {name: 'Thursday', color: '#7090e0'}
      {name: 'Friday', color: '#80f0c0'}
    ]

buildWidget = (element, series, dataURL, widget) ->
  builtWidget = widget.builder(element, series, dataURL, widget)

#activeUsers = buildWidget(document.getElementById('active-users'), ACTIVE_USERS_URL, WStat.widgets.TABLE_PLAIN)
activeDays = buildWidget(document.querySelector('#active-days'), daysSeries, ACTIVE_DAYS_URL, WStat.widgets.GRAPH_LINE)
