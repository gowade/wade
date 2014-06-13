rivets.configure({
  prefix: 'rv'
});

function createObj() {
	return {
		"Username": ":D:D",
		"Password": "kkk"
	}
}

function WadePageManager(startPage, basePath) {
    this.router = new RouteRecognizer();
    this.currentPage = null;
    this.pageHandlers = {};
	this.startPage = startPage;
	this.basePath = basePath;
	this.pages = {};
	this.notFoundPage = "";
	this.pageModels = {};
}

WadePageManager.prototype = {
	cutPath: function(path) {
		if (path.indexOf(this.basePath) === 0) {
			path = path.substring(this.basePath.length, path.length);
		}
		return path
	},
	pathForPage: function(pageId) {
		var page = this.pages[pageId];
		if (page === undefined) {
			throw "no such page #"+pageId+" registered.";
		}
		return page.path;
	},
	setNotFoundPage: function(pageId) {
		this.notFoundPage = this.pathForPage(pageId);
	},
	setPageOnLoad: function() {
		var path = this.cutPath(document.location.pathname);
		path = (path==="/") ? this.pathForPage(this.startPage) : path;
		this.updatePage(path);	
		history.replaceState(null, null, this.basePath+path);
	},
    init: function() {
        var master = this;
        //$(document).ready(function() {
            // looking for all the links and hang on the event, all references in this document
            $("a").on('click', function() {
                var href = $(this).attr("href");
                if (!href || href.indexOf(":") !== 0) {
                    return true;
                }
                //if (href.indexOf("/") !== 0) {
                //     href = $gCurrentPage.data("route") + "/" + href;
                //}
                // keep the link in the browser history
				var pageId = href.substring(1, href.length);
				var path = master.pathForPage(pageId);
				history.pushState(null, master.pages[pageId].title, master.basePath+path);
                try {
					master.updatePage(path);
				} catch (error) {
					console.error(error);
				}

                //
                return false;
            });

            // hang on popstate event triggered by pressing back/forward in browser
            $(window).on('popstate', function(e) {
                
                var returnLocation = history.location || document.location;
				
                master.updatePage(returnLocation.pathname);
                return true;
            });

            //!!!!!!!
            //
            $("welement").hide();
            /*$("wpage").each(function() {
                master.setRoute	wade.SetStartingPage("home")ForPage($(this));
            });*/
			
			master.setPageOnLoad();
        //});
    },

    /*setRouteForPage: function(path, pageElem) {
        var parent = pageElem.parent("wpage");
        var proute = "";
        if (parent.length !== 0) {
            proute = parent.data("route");
            if (proute === undefined) {
                proute = master.setRouteForPage(parent);
            }
        }
        var nroute = proute + (proute ? "/" : "") + pageElem.attr("page");
		
    },*/

    updatePage: function (href) {
		href = this.cutPath(href);
        var matches = this.router.recognize(href);
		console.log("path: "+href);
        if (matches.length === 0) {
			if (this.notFoundPage != "") {
				this.updatePage(this.notFoundPage);
			} else {
				throw "Page not found.";
			}
        }
        var pageId = matches[0].handler();
		var pageElem = $("#"+pageId);
		$("title").text(this.pages[pageId].title);
		if (this.currentPage) {
	        /*if (page.closest(this.currentPage).length === 0) {
	            this.currentPage.hide();
	            page.parents().has(this.currentPage).first() //get common ancestor
	            .children().has(this.currentPage).first().hide(); //hide currentPage's appropriate parent
	        }*/
			this.currentPage.hide();
			this.currentPage.parents("wpage").each(function() {
				$(this).hide();
			});
		}
		
		//var lparent = page;
		pageElem.parents("wpage").each(function() {
			$(this).show();
			/*th.children("wpage").each(function() {
				if (this !== lparent.get(0)) {
					//$(this).hide();
				}
			});
			lparent = th;*/
		});

        pageElem.show();
        this.currentPage = pageElem;
		 
		var handlers = this.pageHandlers[pageId];
		if (handlers !== undefined) {
			for (var i in handlers) {
	            var model = handlers[i]();
				//console.log(model);
	            this.pageModels = rivets.bind(pageElem, model).models;
				console.log(this.pageModels);
			}
		}
    },

    registerHandler: function (pageId, handlerFn) {
		if (this.pageHandlers[pageId] === undefined) {
			this.pageHandlers[pageId] = [];
		}
        this.pageHandlers[pageId].push(handlerFn);
    },
	
	registerPages: function(pages) {
		for (var path in pages) {
			var pageId = pages[path];
			if (this.pages[pageId] != undefined) {
				throw "Page #"+pageId+" has already been registered.";
			}
			var pageElem = $("#"+pageId);
			if (!pageElem.length) {
				throw "There is no such page element #"+pageId+".";
			}
			
			(function(pageId, router, path) {
				router.add([{ path: path, handler: function() {
	            	return pageId;
	        	} }]);
			})(pageId, this.router, path);
		
			this.pages[pageId] = {path: path, title: pageElem.attr("title")};
		}
	}
};

function Wade(startPage, basePath) {
    this.sign = "1'M_7763_W4D3,_817C76!";
    this.pageMan = new WadePageManager(startPage, basePath);
    this._attrPrefix = "attr-";
	this.elements = {};
}

Wade.prototype = {
    registerElement: function(tagid, model) {
        var te = $("#"+tagid);
        if (te.length === 0) {
            throw "Such welement does not exist.";
        }
        if (te.prop("tagName") !== "WELEMENT") {
            throw "The registered `"+tagid+"` is not a welement!";
        }
		this.elements[tagid] = model;
    },
    
    start: function() {
		var wade = this;
		$(document).ready(function() {
			$("wpage").hide();
	        wade.pageMan.init();
			
			for (var tagid in wade.elements) {
				wade.bind(tagid, wade.elements[tagid]);
			}
		});
    },
	
	bind: function(tagid, model) {
		var wade = this;
		var te = $("#"+tagid);
		var attrPrefix = this._attrPrefix;
		var bindPrefix = "bind-";
		var publicAttrs = te.attr("attributes").split(" ");
		/*for (var i in publicAttrs) {
			var attr = publicAttrs[i];
			if (attr.indexOf(attrPrefix) !== 0) {
                throw tagid+": Custom element attribute must have prefix `"+attrPrefix+"`!";
            }
            publicAttrs[i] = attr.substring(attrPrefix.length, attr.length);
		}*/
        for (var i in publicAttrs) {
			var attr = publicAttrs[i];
            if (model[attr] === undefined) {
                throw "Attribute `"+attr+"`is not available in the element model for `"+tagid+"`.";
            }
        }

        var elems = $(tagid);
        elems.each(function() {
            var elem = $(this);
            var mclone = $.extend({}, model);
            for (var i in publicAttrs) {
				var attr = publicAttrs[i];
                var val = elem.attr(attrPrefix+attr);
                if (val !== undefined) {
                    mclone[attr] = val;
                }
				val = elem.attr(bindPrefix+attrPrefix+attr);
				if (val !== undefined) {
					var pageModels = wade.pageMan.pageModels;
					for (i in pageModels) {
						var attrModel = pageModels[i][attr];
						if (attrModel != undefined) {
							mclone[attr] = function() {
								return [attrModel.Username, attrModel.Password];
							}
							break;
						}
					}
				}
            }
			/*if (mclone.Errors !== undefined) {
				mclone["Errors"].Username = function() {
					return "Du`";
				}
				console.log("(");
				console.log(mclone);
				console.log(")");
			}*/
            elem.append(te.html());
            setTimeout(function() {
				//console.log(mclone);
                rivets.bind(elem, mclone);
            }, 20);
        });
	}
};