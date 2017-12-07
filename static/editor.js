Survey.Survey.cssType = "bootstrap";
Survey.defaultBootstrapCss.navigationButton = "btn btn-green";

var surveyUID = $.url().param("survey");

$.get("surveys/" + surveyUID, function(data) {
  editor.text = JSON.stringify(data);
});

var editorOptions = {};
var editor = new SurveyEditor.SurveyEditor("editorElement", editorOptions);

editor.saveSurveyFunc = function() {
  $.ajax({
    method: "PUT",
    url: "surveys/" + surveyUID,
    contentType: "application/json; charset=utf-8",
    data: JSON.stringify(JSON.parse(editor.text))
  })
    .done(function(data) {
      alert("saved");
    })
    .fail(function() {
      alert("failed to save");
    });
};
