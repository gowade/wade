rivets.configure({
  prefix: 'w'
});

function WadePageManager() {
    this.router = new RouteRecognizer();
    this.currentPage = null;
}

WadePageManager.prototype = {
    init: function() {
        var master = this;
        $(document).ready(function() {
            // looking for all the links and hang on the event, all references in this document
            $("a").on('click', function() {
                var href = $(this).attr("href");
                if (href !== "" && href.indexOf("http://") === 0) {
                    return true;
                }
                // if (href.indexOf("/") !== 0) {
                //     href = $gCurrentPage.data("route") + "/" + href;
                // }
                // keep the link in the browser history
                history.pushState(null, null, href);

                master.updatePage(href);

                // here can cause data loading, etc.
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
        //console.log(nroute);
        if (pageElem.attr("page") !== "home") {
            pageElem.hide();
        } else {
            this.currentPage = pageElem;
        }
        return pageElem.data("route", nroute);
    },

    updatePage: function (href) {
        var matches = this.router.recognize(href);
        //console.log(href);
        if (!matches) {
            return;
        }
        var page = matches[0].handler();
        if (page.closest(this.currentPage).length === 0) {
            this.currentPage.hide();
            page.parents().has(this.currentPage).first() //get common ancestor
            .children().has(this.currentPage).first().hide(); //hide currentPage's appropriate parent
        }
        page.show();
        this.currentPage = page;

        return;
    }
};

function Wade() {
    this.pageMan = new WadePageManager();
    this.pageMan.init();
}

Wade.prototype = {
    register: function(tagid, model) {
        var te = $("#"+tagid);
        if (te.length === 0) {
            throw "Such welement does not exist.";
        }
        if (te.prop("tagName") !== "WELEMENT") {
            throw "The registered thing is not a welement!";
        }

        var elems = $(tagid);
        elems.each(function() {
            var elem = $(this);
            var mclone = $.extend({}, model);
            for (var key in mclone) {
                if (model[key] !== undefined && typeof(model[key]) !== "function") {
                    var val = elem.attr("v-"+key);
                    if (val !== undefined) {
                        mclone[key] = val;
                    }
                }
            }
            elem.append(te.html());
            setTimeout(function() {
                rivets.bind(elem, mclone);
            }, 100);
        });
    },
};