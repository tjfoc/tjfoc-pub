%{
package fate
%}

%union{
    val int
    sval string
}

%token <sval> SET, GET, DATA

%%

program: statementlist;

statementlist: statement
            | statementlist ';' statement
            ;
statement:
   | SET DATA DATA {ft.Set($2, $3)}
   | GET DATA { ft.Get($2)}
   ;

%%
