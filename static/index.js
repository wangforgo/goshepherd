const $ = layui.jquery;

layui.use('form', function () {
	const form = layui.form;
	form.on('select(toolSelect)', function (data) {
		const path2Input = $("#path2Input");
		if (data.value === "2") {
			if (path2Input.hasClass("hidden") === true) {
				path2Input.removeClass("hidden");
				$("#path2").attr("lay-verify", "required").val("");
				form.render();
			}
		} else {
			if (path2Input.hasClass("hidden") === false) {
				path2Input.addClass("hidden");
				$("#path2").attr("lay-verify", "").val("dummy-placeholder");
				form.render();
			}
		}
	});

	// submit form
	form.on('submit(myForm)', function (data) {
		if ($("#openBtn").hasClass("layui-btn-disabled")) {
			return false;
		}
		$("#openBtn").addClass("layui-btn-disabled");

		toolType = $("#tool").val();
		prjName = $("#projectName").val();
		path1 = $("#path1").val();
		path2 = $("#path2").val() ==="dummy-placeholder" ? "" : $("#path2").val()
		$.ajax({
			type: "get",
			url: "http://127.0.0.1:7777/api",
			data: {
				op: "add",
				tool: toolType,
				path1: path1,
				path2: path2
			},
			success: function (data) {
				if (data === "") {
					alertCheckGoShepherd();
					return false;
				}

				var port = Number(data);
				if (isNaN(port)) {
					layer.alert("err", {
						type: 0,
						title: `Warning`,
						content: data
					})
					return false;
				}

				pathNode = '<td><p>'+path1+'</p>'
				if (path2 !== "") {
					pathNode += '<p>'+path2+'</p>'
				}
				pathNode += '</td>'

				$("#tableBody").append('<tr>\n' +
					'<td>' + prjName + '</td>\n' +
					'<td><a style="color:#009688" href="http://127.0.0.1:'+port+'" target="_blank">http://127.0.0.1:'+port+'</a></td>\n' +
					pathNode+
					'<td>\n' +
					' <button type="button" port='+port+' class="delBtn layui-btn layui-btn-sm layui-btn-danger">\n' +
					' <i class="layui-icon">&#xe640;</i>\n' +
					' </button>\n' +
					'</td>\n' +
					'</tr>');

				return false;
			},
			error: function (data) {
				alertCheckGoShepherd();
				return false;
			}
		});

		$("#openBtn").removeClass("layui-btn-disabled");
		return false;
	});

	// delete item
	$("#tableBody").on("click",".delBtn", function () {
		$.ajax({
			type: "get",
			url: "http://127.0.0.1:7777/api",
			data: {
				op: "rmv",
				port: $(this).attr("port"),
			}
		})
		$(this).parent().parent().remove();
	});
});

function alertCheckGoShepherd() {
	var layer = layui.layer;
	layer.alert("err", {
		type: 0,
		title: `Warning`,
		content: `Oops...Seems that GoShepherd is not working well...Please check it!`
	})
}




