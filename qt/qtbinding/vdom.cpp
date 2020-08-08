#include "vdom.hpp"


void Node::diff(DeltaNotifier* ctx, Node* parent, Node* old, Node* _new) {
    assert(ctx != nullptr);
    assert(!(old == nullptr && _new == nullptr));
    auto parent_id = reinterpret_cast<uintptr>(parent);
    auto old_id = reinterpret_cast<uintptr>(old);
    auto new_id = reinterpret_cast<uintptr>(_new);
    if (old == _new) { return; }
    if (old == nullptr) {
        ctx->AppendNode(parent_id, new_id, _new->tagName);
    } else if (_new == nullptr) {
        ctx->RemoveNode(parent_id, old_id);
        old->deleteLater();
    } else {
        if (old->tagName == _new->tagName) {
            ctx->UpdateNode(old_id, new_id);
        } else {
            ctx->ReplaceNode(old_id, new_id, _new->tagName);
        }
    }
    if (_new != nullptr) {
        auto id = new_id;
        auto node = _new;
        auto& new_style = _new->style;
        if (old != nullptr) {
            auto& old_style = old->style;
            for (auto e = old_style.begin(); e != old_style.end(); e++) {
                auto old_key = e.key();
                if (!(new_style.contains(old_key))) {
                    ctx->EraseStyle(id, old_key);
                }
            }
        }
        for (auto e = new_style.begin(); e != new_style.end(); e++) {
            ctx->ApplyStyle(id, e.key(), e.value());
        }
        auto& new_events = _new->events;
        if (old != nullptr) {
            auto& old_events = old->events;
            for (auto e = old_events.begin(); e != old_events.end(); e++) {
                auto old_key = e.key();
                auto old_opts = e.value();
                if (new_events.contains(old_key)) {
                    auto new_opts = new_events[old_key];
                    if (new_opts != old_opts) {
                        ctx->DetachEvent(id, old_key, false);
                        delete old_opts;
                        ctx->AttachEvent(id, old_key,
                            new_opts->prevent, new_opts->stop, new_opts->handler);
                    }
                } else {
                    ctx->DetachEvent(id, old_key, true);
                    delete old_opts;
                }
            }
        }
        for (auto e = new_events.begin(); e != new_events.end(); e++) {
            auto new_key = e.key();
            if (old != nullptr) {
                auto& old_events = old->events;
                if (!(old_events.contains(new_key))) {
                    auto new_opts = e.value();
                    ctx->AttachEvent(id, new_key,
                        new_opts->prevent, new_opts->stop, new_opts->handler);
                }
            } else {
                auto new_opts = e.value();
                ctx->AttachEvent(id, new_key,
                    new_opts->prevent, new_opts->stop, new_opts->handler);
            }
        }
        auto& new_children = _new->children;
        int i = 0;
        for (auto child: new_children) {
            Node* old_child = nullptr;
            if (old != nullptr) {
                auto& old_children = old->children;
                if (i < old_children.length()) {
                    old_child = old_children[i];
                }
            }
            diff(ctx, node, old_child, child);
            i += 1;
        }
        if (old != nullptr) {
            auto& old_children = old->children;
            int old_len = old_children.length();
            int new_len = new_children.length();
            if (old_len > new_len) {
                for (int i = new_len; i < old_len; i += 1) {
                    Node* old_child = old_children[i];
                    diff(ctx, node, old_child, nullptr);
                }
            }
        }
    }
};
