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
				$("#path2").attr("lay-verify", "").val("dummy");
				form.render();
			}
		}
	});

	// 监听提交
	form.on('submit(myForm)', function (data) {
		if ($("#openBtn").hasClass("layui-btn-disabled")) {
			return false;
		}
		$("#openBtn").addClass("layui-btn-disabled");

		toolType = $("#tool").val();
		prjName = $("#projectName").val();
		path1 = $("#path1").val();
		path2 = $("#path2").val();

		$.ajax({
			type: "get",
			url: "http://localhost:7777/api",
			data: {
				op: "add",
				tool: toolType,
				path1: path1,
				path2: path2
			},
			success: function (data) {
				if (data === "") {
					layer.alert("err", {
						type: 0,
						title: `Warning`,
						content: "please check goshepherd...",
						btn: `ok`,
					});
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


				path = path1
				if (path2 !== "") {
					path = path1 + path2 // todo: newline for path2
				}

				$("#tableBody").append('<tr>\n' +
					'<td>' + prjName + '</td>\n' +
					'<td><a href="http://localhost:'+port+'" target="_blank">http://localhost:'+port+'</a></td>\n' +
					'<td>' + path + '</td>\n' +
					'<td>\n' +
					' <button type="button" class="delBtn layui-btn layui-btn-sm layui-btn-danger">\n' +
					' <i class="layui-icon">&#xe640;</i>\n' +
					' </button>\n' +
					'</td>\n' +
					'</tr>');

				return false;
			},
			error: function (data) {
				var layer = layui.layer;
				layer.alert("err", {
					type: 0,
					title: `Warning`,
					content: `something is not working well`
				})
				return false;
			}
		});

		$("#openBtn").removeClass("layui-btn-disabled");
		return false;
	});

	// delete item
	$("#tableBody").on("click",".delBtn", function () {
		console.log("delete...");
		$(this).parent().parent().remove();
	});
});






