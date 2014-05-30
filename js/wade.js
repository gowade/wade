rivets.configure({
  prefix: 'w'
});

function WadePageManager(startPage, basePath) {
    this.router = new RouteRecognizer();
    this.currentPage = null;
    this.pageHandlers = {};
	this.startPage = startPage;
	this.basePath = basePath;
}

WadePageManager.prototype = {
	cutPath: function(path) {
		if (path.indexOf(this.basePath) === 0) {
			path = path.substring(this.basePath.length, path.length);
		}
		return path
	},
	setStartingPage: function() {
		var path = this.cutPath(document.location.pathname);
		path = (path==="/") ? this.startPage : path;
		this.updatePage(path);	
		history.replaceState(null, null, this.basePath+path);
		if (!this.currentPage) {
			throw "Starting page `"+this.startPage+"` not found.";
		}
	},
    init: function() {
        var master = this;
        $(document).ready(function() {
            // looking for all the links and hang on the event, all references in this document
            $("a").on('click', function() {
                var href = $(this).attr("href");
                if (href !== "" && href.indexOf("http://") === 0) {
                    return true;
                }
                //if (href.indexOf("/") !== 0) {
                //     href = $gCurrentPage.data("route") + "/" + href;
                //}
                // keep the link in the browser history
				history.pushState(null, null, href);
                try {
					master.updatePage(href);
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
            $("wpage").each(function() {
                master.setRouteForPage($(this));
            });
			
			master.setStartingPage();
        });
    },

    setRouteForPage: function(pageElem) {
        var parent = pageElem.parent("wpage");
        var proute = "";
        if (parent.length !== 0) {
            proute = parent.data("route");
            if (proute === undefined) {
                proute = master.setRouteForPage(parent);
            }
        }
        var nroute = proute + (proute ? "~" : "") + pageElem.attr("page");
        this.router.add([{ path: nroute, handler: function() {
            return pageElem;
        } }]);
		
		pageElem.hide();
        return pageElem.data("route", nroute);
    },

    updatePage: function (href) {
		href = this.cutPath(href);
        var matches = this.router.recognize(href);
		console.log(href);
        if (!matches) {
            return;
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
		
		var lparent = page;
		page.parents("wpage").each(function() {
			console.log("^^");
			var th = $(this);
			th.show();
			th.children("wpage").each(function() {
				if (this !== lparent.get(0)) {
					$(this).hide();
				}
			});
			lparent = th;
		});

        page.show();
        this.currentPage = page;
        var handler = page.attr("handler");
        if (handler !== undefined) {
            if (this.pageHandlers[handler] === undefined) {
                throw "Page handler `"+handler+"` has not been registered.";
            }
            var model = this.pageHandlers[handler]();
            rivets.bind(page, model);
        }
    },

    registerHandler: function (name, handlerFn) {
        this.pageHandlers[name] = handlerFn;
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
					console.log(attr+":"+val);
                    mclone[attr] = val;
                }
            }
            elem.append(te.html());
            setTimeout(function() {
                rivets.bind(elem, mclone);
            }, 20);
        });
    },

    registerPageHandler: function(name, fn) {
        this.pageMan.registerHandler(name, fn);
    },
    
    start: function() {
        this.pageMan.init();
    }
};