// FizzBuzz implementation
function fizzBuzz(n = 100) {
    console.log('FizzBuzz from 1 to', n);
    console.log('==================');
    
    for (let i = 1; i <= n; i++) {
        let output = '';
        
        if (i % 3 === 0) output += 'da fizz';
        if (i % 5 === 0) output += 'dat buzz';
        
        console.log(output || i);
    }
}

// Execute FizzBuzz
fizzBuzz(15);