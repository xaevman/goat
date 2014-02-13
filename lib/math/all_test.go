//  ---------------------------------------------------------------------------
//
//  all_test.go
//
//  Copyright (c) 2014, Jared Chavez. 
//  All rights reserved.
//
//  Use of this source code is governed by a BSD-style
//  license that can be found in the LICENSE file.
//
//  -----------

package math

// Stdlib imports.
import(
    "fmt"
    "testing"
)

// KahanSum test data.
var kahanSeries = []float64 {
    -7949007.7765421,
    -5199414.0144435,
    -4291398.7088443,
    -3128598.5247831,
    2054702.9339032,
    6787782.3099437,
    2955415.9347692,
    7280850.111172,
    7461681.8863254,
    -589385.74539003,
    6361060.8672542,
    -9775664.9552731,
    -4367239.1746041,
    6086489.3980727,
    4360032.1627967,
    4508700.1354008,
    3690154.074547,
    5144006.4400174,
    -1579990.4296081,
    -8241152.4924641,
    -7459892.8901646,
    2597973.3153237,
    -6884611.0985077,
    -6886788.7076395,
    3390125.089041,
    -8274028.3190617,
    8453052.6206144,
    8240321.1753072,
    1221807.2690171,
    7212656.3439205,
    -6639193.7884685,
    -7422158.8752336,
    -5055273.9273083,
    -2153998.0136575,
    9308815.6074792,
    2929914.534525,
    -8490178.3142659,
    -5510705.0740676,
    42938.566786675,
    3241735.5260075,
    -8310228.5760968,
    770721.6640798,
    -5576605.5153574,
    717372.69438681,
    -7619913.4055618,
    -6790661.4378051,
    9236646.7226467,
    -4602194.8822784,
    -8675034.0548693,
    7203619.2460002,
}

// KahanSum test result.
var kahanResult = -3.0214742072958037e+07

// TextKahanSum adds 50 floats with a wide range of values and checks results.
func TestKahanSum(t *testing.T) {
    fmt.Println()

    ks := new(KahanSum)

    for i := 0; i < len(kahanSeries); i++ {
        ks.Add(kahanSeries[i])
    }

    if ks.Sum() != kahanResult {
        t.Fatalf("Bad sum: %v != %v", ks.Sum(), kahanResult)
    }

    fmt.Printf("With KahanSum: %v\n", ks.Sum())

    var sum float64
    for i := 0; i < len(kahanSeries); i++ {
        sum += kahanSeries[i]
    }
    fmt.Printf("Std sum:       %v\n", sum)

    fmt.Println()
}
