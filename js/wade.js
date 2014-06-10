rivets.configure({
  prefix: 'w'
});

function WadePageManager(startPage, basePath) {
    this.router = new RouteRecognizer();
    this.currentPage = null;
    this.pageHandlers = {};
	this.startPage = startPage;
	this.basePath = basePath;
	this.pages = {};
	this.notFoundPage = "";
}

WadePageManager.prototype = {
	cutPath: function(path) {
		if (path.indexOf(this.basePath) === 0) {
			path = path.substring(this.basePath.length, path.length);
		}
		return path
	},
	pathForPage: function(pageId) {
		var path = this.pages[pageId];
		if (path === undefined) {
			throw "no such page #"+pageId+" registered.";
		}
		return path;
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
        $(document).ready(function() {
            // looking for all the links and hang on the event, all references in this document
            $("a").on('click', function() {
                var href = $(this).attr("href");
                if (href !== "" && href.indexOf(":") !== 0) {
                    return true;
                }
                //if (href.indexOf("/") !== 0) {
                //     href = $gCurrentPage.data("route") + "/" + href;
                //}
                // keep the link in the browser history
				var path = master.pathForPage(href.substring(1, href.length));
				history.pushState(null, null, master.basePath+path);
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
        });
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
			/*if (this.notFoundPage != "") {
				this.updatePage(this.notFoundPage);
			} else {
				throw "Page not found.";
			}*/
        }
        var page = matches[0].handler();
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
		page.parents("wpage").each(function() {
			$(this).show();
			/*th.children("wpage").each(function() {
				if (this !== lparent.get(0)) {
					//$(this).hide();
				}
			});
			lparent = th;*/
		});

        page.show();
        this.currentPage = page;
		var handlers = this.pageHandlers[page.attr("id")];
		if (handlers !== undefined) {
			for (var i in handlers) {
	            var model = handlers[i]();
	            rivets.bind(page, model);
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
			
			(function(page, router, path) {
				router.add([{ path: path, handler: function() {
	            	return page;
	        	} }]);
			})(pageElem, this.router, path);
		
			this.pages[pageId] = path;
		}
	}
};

function Wade(startPage, basePath) {
    this.sign = "1'M_7763_W4D3,_817C76!";
    this.pageMan = new WadePageManager(startPage, basePath);
    this._attrPrefix = "attr-";
}

Wade.prototype = {
    register: function(tagid, model) {
        var te = $("#"+tagid);
		var attrPrefix = this._attrPrefix;
        if (te.length === 0) {
            throw "Such welement does not exist.";
        }
        if (te.prop("tagName") !== "WELEMENT") {
            throw "The registered `"+tagid+"` is not a welement!";
        }

        var publicAttrs = te.attr("attributes").split(" ");
		for (var i in publicAttrs) {
			var attr = publicAttrs[i];
			if (attr.indexOf(attrPrefix) !== 0) {
                throw tagid+": Custom element attribute must have prefix `"+attrPrefix+"`!";
            }
            publicAttrs[i] = attr.substring(attrPrefix.length, attr.length);
		}
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
            }
            elem.append(te.html());
            setTimeout(function() {
                rivets.bind(elem, mclone);
            }, 20);
        });
    },
    
    start: function() {
		$("wpage").hide();
        this.pageMan.init();
    }
};