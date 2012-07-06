# Public Domain (-) 2012 The Wifistat Authors.
# See the Wifistat UNLICENSE file for details.

define 'wifistat', (exports, root) ->

  Rickshaw = root.Rickshaw

  widgetSize =
    DEFAULT:
      width: '100%'
      height: '250px'

  exports.widgets =
    GRAPH_LINE: 
      type: Rickshaw.Graph.Ajax
      renderer: 'line'
    GRAPH_BAR: 
      type: Rickshaw.Graph.Ajax
      renderer: 'bar'

  exports.buildWidget = (element, dataURL, widget) ->
    graph = new widget.type
      element: element
      width: widget.width
      height: widget.height
      renderer: widget.renderer
      dataURL: dataURL


