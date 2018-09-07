#include "compile.h"
#include <cstring>

static char injectGasFormat[] =
    "(function(){\n"
    "const source = \"%s\";\n"
    "return injectGas(source);\n"
    "})();";

int compile(SandboxPtr ptr, const char *code, const char **compiledCode) {
    Sandbox *sbx = static_cast<Sandbox*>(ptr);
    Isolate *isolate = sbx->isolate;

    Locker locker(isolate);
    Isolate::Scope isolate_scope(isolate);
    HandleScope handle_scope(isolate);

    Local<Context> context = sbx->context.Get(isolate);
    Context::Scope context_scope(context);

    char *injectCode = nullptr;
    asprintf(&injectCode, injectGasFormat, code);

    Local<String> source = String::NewFromUtf8(isolate, injectCode, NewStringType::kNormal).ToLocalChecked();
    free(injectCode);
    Local<String> fileName = String::NewFromUtf8(isolate, "__inject_ga.js", NewStringType::kNormal).ToLocalChecked();
    Local<Script> script = Script::Compile(source, fileName);

    if (!script.IsEmpty()) {
        Local<Value> result = script->Run();
        if (!result.IsEmpty()) {
            String::Utf8Value retStr(result);
            *compiledCode = strdup(*retStr);
            return 0;
        }
    }
    return 1;
}
